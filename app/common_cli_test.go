package app

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdkKeys "github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/tests"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	tmCTypes "github.com/tendermint/tendermint/rpc/core/types"
	tmRPCTypes "github.com/tendermint/tendermint/rpc/lib/types"
	tmTypes "github.com/tendermint/tendermint/types"

	"github.com/dfinance/dnode/cmd/config"
	dnConfig "github.com/dfinance/dnode/cmd/config"
	ccTypes "github.com/dfinance/dnode/x/currencies/types"
	msTypes "github.com/dfinance/dnode/x/multisig/types"
	"github.com/dfinance/dnode/x/oracle"
	poaTypes "github.com/dfinance/dnode/x/poa/types"
)

type CLITester struct {
	RootDir           string
	ChainID           string
	MonikerID         string
	AccountPassphrase string
	Accounts          map[string]*CLIAccount
	Cdc               *codec.Codec
	t                 *testing.T
	validatorAddrs    []string
	wbdBinary         string
	wbcliBinary       string
	rpcAddress        string
	rpcPort           string
	p2pAddress        string
	vmConnectAddress  string
	vmListenAddress   string
	daemon            *CLICmd
}

func NewCLITester(t *testing.T) *CLITester {
	RequireCliTestEnv(t)

	sdkConfig := sdk.GetConfig()
	dnConfig.InitBechPrefixes(sdkConfig)

	_, vmConnectPort, err := server.FreeTCPAddr()
	require.NoError(t, err, "FreeTCPAddr for VM connect")
	_, vmListenPort, err := server.FreeTCPAddr()
	require.NoError(t, err, "FreeTCPAddr for VM listen")
	srvAddr, srvPort, err := server.FreeTCPAddr()
	require.NoError(t, err, "FreeTCPAddr for srv")
	p2pAddr, _, err := server.FreeTCPAddr()
	require.NoError(t, err, "FreeTCPAddr for p2p")

	cdc := MakeCodec()
	ct := CLITester{
		t:                 t,
		Cdc:               cdc,
		wbdBinary:         "dnode",
		wbcliBinary:       "dncli",
		ChainID:           "test-chain",
		MonikerID:         "test-moniker",
		AccountPassphrase: "passphrase",
		rpcAddress:        srvAddr,
		rpcPort:           srvPort,
		p2pAddress:        p2pAddr,
		vmConnectAddress:  fmt.Sprintf("127.0.0.1:%s", vmConnectPort),
		vmListenAddress:   fmt.Sprintf("127.0.0.1:%s", vmListenPort),
		Accounts:          make(map[string]*CLIAccount, 0),
	}

	smallAmount, ok := sdk.NewIntFromString("5000000000000")
	require.True(t, ok, "NewInt for smallAmount")
	bigAmount, ok := sdk.NewIntFromString("90000000000000000000000000")
	require.True(t, ok, "NewInt for bigAmount")

	ct.Accounts["pos"] = &CLIAccount{
		Coins: map[string]sdk.Coin{
			config.MainDenom: sdk.NewCoin(config.MainDenom, bigAmount),
		},
	}
	ct.Accounts["bank"] = &CLIAccount{
		Coins: map[string]sdk.Coin{
			config.MainDenom: sdk.NewCoin(config.MainDenom, bigAmount),
		},
	}
	ct.Accounts["validator1"] = &CLIAccount{
		Coins: map[string]sdk.Coin{
			config.MainDenom: sdk.NewCoin(config.MainDenom, smallAmount),
		},
		IsPOAValidator: true,
	}
	ct.Accounts["validator2"] = &CLIAccount{
		Coins: map[string]sdk.Coin{
			config.MainDenom: sdk.NewCoin(config.MainDenom, smallAmount),
		},
		IsPOAValidator: true,
	}
	ct.Accounts["validator3"] = &CLIAccount{
		Coins: map[string]sdk.Coin{
			config.MainDenom: sdk.NewCoin(config.MainDenom, smallAmount),
		},
		IsPOAValidator: true,
	}
	ct.Accounts["oracle1"] = &CLIAccount{
		Coins: map[string]sdk.Coin{
			config.MainDenom: sdk.NewCoin(config.MainDenom, smallAmount),
		},
		IsOracleNominee: true,
	}
	ct.Accounts["oracle2"] = &CLIAccount{
		Coins: map[string]sdk.Coin{
			config.MainDenom: sdk.NewCoin(config.MainDenom, smallAmount),
		},
		IsOracleNominee: false,
	}
	ct.Accounts["oracle3"] = &CLIAccount{
		Coins: map[string]sdk.Coin{
			config.MainDenom: sdk.NewCoin(config.MainDenom, smallAmount),
		},
		IsOracleNominee: false,
	}

	rootDir, err := ioutil.TempDir("/tmp", fmt.Sprintf("wd-cli-test-%s-", ct.t.Name()))
	require.NoError(t, err, "TempDir")
	ct.RootDir = rootDir

	ct.initChain()

	ct.startDemon()

	ct.UpdateAccountsBalance()

	return &ct
}

func (ct *CLITester) Close() {
	if ct.daemon != nil {
		ct.daemon.Stop()
	}

	if ct.RootDir != "" {
		os.RemoveAll(ct.RootDir)
	}
}

func (ct *CLITester) newWbdCmd() *CLICmd {
	cmd := &CLICmd{t: ct.t, base: ct.wbdBinary}
	cmd.AddArg("home", ct.RootDir)

	return cmd
}

func (ct *CLITester) newWbcliCmd() *CLICmd {
	cmd := &CLICmd{t: ct.t, base: ct.wbcliBinary}
	cmd.AddArg("home", ct.RootDir)
	cmd.AddArg("chain-id", ct.ChainID)
	cmd.AddArg("output", "json")

	return cmd
}

func (ct *CLITester) newTxRequest() *TxRequest {
	return &TxRequest{
		t:              ct.t,
		cdc:            ct.Cdc,
		cmd:            ct.newWbcliCmd(),
		nodeRpcAddress: ct.rpcAddress,
		accPassphrase:  ct.AccountPassphrase,
	}
}

func (ct *CLITester) newQueryRequest(resultObj interface{}) *QueryRequest {
	return &QueryRequest{
		t:              ct.t,
		cdc:            ct.Cdc,
		cmd:            ct.newWbcliCmd(),
		nodeRpcAddress: ct.rpcAddress,
		resultObj:      resultObj,
	}
}

func (ct *CLITester) initChain() {
	// init chain
	cmd := ct.newWbdCmd().AddArg("", "init").AddArg("", ct.MonikerID).AddArg("chain-id", ct.ChainID)
	cmd.CheckSuccessfulExecute(nil)

	// adjust tendermint config (make blocks generation faster)
	{
		cfgFile := path.Join(ct.RootDir, "config", "config.toml")
		_, err := os.Stat(cfgFile)
		require.NoError(ct.t, err, "reading config.toml file")
		viper.SetConfigFile(cfgFile)
		require.NoError(ct.t, viper.ReadInConfig())

		viper.Set("consensus.timeout_propose", "250ms")
		viper.Set("consensus.timeout_propose_delta", "250ms")
		viper.Set("consensus.timeout_prevote", "250ms")
		viper.Set("consensus.timeout_prevote_delta", "250ms")
		viper.Set("consensus.timeout_precommit", "250ms")
		viper.Set("consensus.timeout_precommit_delta", "250ms")
		viper.Set("consensus.timeout_commit", "250ms")

		require.NoError(ct.t, viper.WriteConfig(), "saving config.toml file")
	}

	// configure accounts
	{
		poaValidatorIdx := 0
		for accName, accValue := range ct.Accounts {
			// create key
			{
				cmd := ct.newWbcliCmd().
					AddArg("", "keys").
					AddArg("", "add").
					AddArg("", accName)
				output := sdkKeys.KeyOutput{}

				cmd.CheckSuccessfulExecute(&output, ct.AccountPassphrase, ct.AccountPassphrase)
				accValue.Name = output.Name
				accValue.Address = output.Address
				accValue.PubKey = output.PubKey
				accValue.Mnemonic = output.Mnemonic
			}

			// genesis account
			{
				cmd := ct.newWbdCmd().
					AddArg("", "add-genesis-account").
					AddArg("", accValue.Address)

				require.NotEmpty(ct.t, accValue.Coins, "account %q: no coins", accName)
				var coinsArg []string
				for _, coin := range accValue.Coins {
					coinsArg = append(coinsArg, coin.String())
				}
				cmd.AddArg("", strings.Join(coinsArg, ","))

				cmd.CheckSuccessfulExecute(nil)
			}

			// POA validator
			if accValue.IsPOAValidator {
				require.True(ct.t, poaValidatorIdx < len(ethAddresses), "add more predefined ethAddresses")

				cmd := ct.newWbdCmd().
					AddArg("", "add-genesis-poa-validator").
					AddArg("", accValue.Address).
					AddArg("", ethAddresses[poaValidatorIdx])
				poaValidatorIdx++

				cmd.CheckSuccessfulExecute(nil)

				ct.validatorAddrs = append(ct.validatorAddrs, accValue.Address)
			}

			// Oracle nominee
			if accValue.IsOracleNominee {
				cmd := ct.newWbdCmd().
					AddArg("", "add-oracle-nominees-gen").
					AddArg("", accValue.Address)

				cmd.CheckSuccessfulExecute(nil)
			}
		}
	}

	// validator genTX
	{
		cmd := ct.newWbdCmd().
			AddArg("", "gentx").
			AddArg("home-client", ct.RootDir).
			AddArg("name", "pos").
			AddArg("amount", ct.Accounts["pos"].Coins[config.MainDenom].String())

		cmd.CheckSuccessfulExecute(nil, ct.AccountPassphrase)
	}

	// change default genesis params
	{
		appState := ct.genesisState()

		// staking default denom change
		stakingGenesis := staking.GenesisState{}
		require.NoError(ct.t, ct.Cdc.UnmarshalJSON(appState["staking"], &stakingGenesis), "unmarshal staking genesisState")

		stakingGenesis.Params.BondDenom = config.MainDenom
		stakingGenesisRaw, err := ct.Cdc.MarshalJSON(stakingGenesis)
		require.NoError(ct.t, err, "marshal staking genesisState")
		appState["staking"] = stakingGenesisRaw

		ct.updateGenesisState(appState)
	}

	// collect genTXs
	{
		cmd := ct.newWbdCmd().AddArg("", "collect-gentxs")
		cmd.CheckSuccessfulExecute(nil)
	}

	// validate genesis
	{
		cmd := ct.newWbdCmd().AddArg("", "validate-genesis")
		cmd.CheckSuccessfulExecute(nil)
	}

	// prepare VM config
	{
		vmConfig := dnConfig.DefaultVMConfig()
		vmConfig.Address, vmConfig.DataListen = ct.vmConnectAddress, ct.vmListenAddress
		dnConfig.WriteVMConfig(ct.RootDir, vmConfig)
	}
}

func (ct *CLITester) startDemon() {
	const startRetries = 100

	require.Nil(ct.t, ct.daemon)

	cmd := ct.newWbdCmd().
		AddArg("", "start").
		AddArg("rpc.laddr", ct.rpcAddress).
		AddArg("p2p.laddr", ct.p2pAddress)
	cmd.Start(true)

	// wait for the node to start up
	i := 0
	for ; i < startRetries; i++ {
		time.Sleep(50 * time.Millisecond)
		blockHeight, err := ct.GetCurrentBlockHeight()
		if err != nil {
			continue
		}
		if blockHeight < 2 {
			continue
		}

		break
	}
	if i == startRetries {
		ct.t.Fatalf("wait for the node to start up: failed")
	}
}

func (ct *CLITester) updateGenesisState(appState GenesisState) {
	genesisFile := server.NewDefaultContext().Config.Genesis
	genesisPath := path.Join(ct.RootDir, genesisFile)

	cdc := codec.New()
	genDoc, err := tmTypes.GenesisDocFromFile(genesisPath)
	require.NoError(ct.t, err, "reading genesis file %q", genesisPath)

	appStateRaw, err := cdc.MarshalJSON(appState)
	require.NoError(ct.t, err, "marshal updated appState")
	genDoc.AppState = appStateRaw

	require.NoError(ct.t, genDoc.SaveAs(genesisPath), "saving updated genesis file %q", genesisPath)
}

func (ct *CLITester) genesisState() GenesisState {
	genesisFile := server.NewDefaultContext().Config.Genesis
	genesisPath := path.Join(ct.RootDir, genesisFile)

	cdc := codec.New()
	genDoc, err := tmTypes.GenesisDocFromFile(genesisPath)
	require.NoError(ct.t, err, "reading genesis file %q", genesisPath)

	appState := GenesisState{}
	require.NoError(ct.t, cdc.UnmarshalJSON(genDoc.AppState, &appState), "unmarshal appState")

	return appState
}

func (ct *CLITester) UpdateAccountsBalance() {
	for accName, prevAcc := range ct.Accounts {
		q, curAcc := ct.QueryAccount(prevAcc.Address)
		q.CheckSucceeded()

		for _, curCoin := range curAcc.Coins {
			doUpdate := false
			prevCoin, ok := prevAcc.Coins[curCoin.Denom]
			if !ok {
				doUpdate = true
				ct.t.Logf("Account %q balance updated: %q %s\n", accName, curCoin.Denom, curCoin.Amount.String())
			} else if !curCoin.Amount.Equal(prevCoin.Amount) {
				doUpdate = true
				ct.t.Logf("Account %q balance updated: %q %s -> %s\n", accName, curCoin.Denom, prevCoin.Amount.String(), curCoin.Amount.String())
			}

			if doUpdate {
				prevAcc.Coins[curCoin.Denom] = curCoin
			}
		}
	}
}

func (ct *CLITester) GetCurrentBlockHeight() (int64, error) {
	url := fmt.Sprintf("http://localhost:%s/block", ct.rpcPort)

	res, err := http.Get(url)
	if err != nil {
		return -1, fmt.Errorf("block request: %w", err)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return -1, fmt.Errorf("body read: %w", err)
	}

	if err := res.Body.Close(); err != nil {
		return -1, fmt.Errorf("body close: %w", err)
	}

	resultResp := tmRPCTypes.RPCResponse{}
	if err := ct.Cdc.UnmarshalJSON(body, &resultResp); err != nil {
		return -1, fmt.Errorf("body unmarshal: %w", err)
	}

	resultBlock := tmCTypes.ResultBlock{}
	if err := ct.Cdc.UnmarshalJSON(resultResp.Result, &resultBlock); err != nil {
		return -1, fmt.Errorf("result unmarshal: %w", err)
	}

	if resultBlock.Block == nil {
		return 0, nil
	}

	return resultBlock.Block.Height, nil
}

func (ct *CLITester) WaitForNextBlocks(n int64) int64 {
	prevHeight, err := ct.GetCurrentBlockHeight()
	require.NoError(ct.t, err, "prevBlockHeight")

	for {
		time.Sleep(time.Millisecond * 5)

		curHeight, err := ct.GetCurrentBlockHeight()
		require.NoError(ct.t, err, "curBlockHeight")
		if curHeight - prevHeight >= n {
			return curHeight
		}
	}
}

func (ct *CLITester) ConfirmCall(uniqueID string) {
	validatorAddrs := append([]string(nil), ct.validatorAddrs...)

	// get callID
	q, call := ct.QueryMultisigUnique(uniqueID)
	q.CheckSucceeded()
	require.Equal(ct.t, uniqueID, call.Call.UniqueID)
	require.False(ct.t, call.Call.Approved)
	require.False(ct.t, call.Call.Rejected)
	require.False(ct.t, call.Call.Executed)
	require.False(ct.t, call.Call.Failed)
	for _, voteAddr := range call.Votes {
		for i, validatorAddr := range validatorAddrs {
			if voteAddr.String() == validatorAddr {
				validatorAddrs = append(validatorAddrs[:i], validatorAddrs[i+1:]...)
			}
		}
	}

	// check multisig minValidators
	requiredVotes := int(poaTypes.DefaultMinValidators)/2 + 1
	availableVotes := len(call.Votes) + len(validatorAddrs)
	require.LessOrEqual(ct.t, requiredVotes, availableVotes, "not enough validators to confirm call")

	// send confirms
	for i := 0; i < requiredVotes-len(call.Votes); i++ {
		ct.TxMultisigConfirmCall(validatorAddrs[i], call.Call.MsgID).CheckSucceeded()
	}
	ct.WaitForNextBlocks(1)

	// check call approved
	{
		q, call := ct.QueryMultisigUnique(uniqueID)
		q.CheckSucceeded()
		require.Equal(ct.t, uniqueID, call.Call.UniqueID)
		require.True(ct.t, call.Call.Approved)
	}
}

func (ct *CLITester) TxCurrenciesIssue(recipientAddr, fromAddr, symbol string, amount sdk.Int, decimals int8, issueID string) *TxRequest {
	r := ct.newTxRequest()
	r.SetCmd(
		"currencies",
		fromAddr,
		"ms-issue-currency",
		symbol,
		amount.String(),
		strconv.Itoa(int(decimals)),
		recipientAddr,
		issueID)

	return r
}

func (ct *CLITester) TxCurrenciesDestroy(recipientAddr, fromAddr, symbol string, amount sdk.Int) *TxRequest {
	r := ct.newTxRequest()
	r.SetCmd(
		"currencies",
		fromAddr,
		"destroy-currency",
		ct.ChainID,
		symbol,
		amount.String(),
		recipientAddr)

	return r
}

func (ct *CLITester) TxOracleAddAsset(nomineeAddress, assetCode string, oracleAddresses ...string) *TxRequest {
	r := ct.newTxRequest()
	r.SetCmd(
		"oracle",
		"",
		"add-asset",
		nomineeAddress,
		assetCode,
		strings.Join(oracleAddresses, ","))

	return r
}

func (ct *CLITester) TxOracleSetAsset(nomineeAddress, assetCode string, oracleAddresses ...string) *TxRequest {
	r := ct.newTxRequest()
	r.SetCmd(
		"oracle",
		"",
		"set-asset",
		nomineeAddress,
		assetCode,
		strings.Join(oracleAddresses, ","))

	return r
}

func (ct *CLITester) TxOracleAddOracle(nomineeAddress, assetCode string, oracleAddress string) *TxRequest {
	r := ct.newTxRequest()
	r.SetCmd(
		"oracle",
		"",
		"add-oracle",
		nomineeAddress,
		assetCode,
		oracleAddress)

	return r
}

func (ct *CLITester) TxOracleSetOracles(nomineeAddress, assetCode string, oracleAddresses ...string) *TxRequest {
	r := ct.newTxRequest()
	r.SetCmd(
		"oracle",
		"",
		"set-oracles",
		nomineeAddress,
		assetCode,
		strings.Join(oracleAddresses, ","))

	return r
}

func (ct *CLITester) TxOraclePostPrice(nomineeAddress, assetCode string, price sdk.Int, receivedAt time.Time) *TxRequest {
	r := ct.newTxRequest()
	r.SetCmd(
		"oracle",
		"",
		"postprice",
		nomineeAddress,
		assetCode,
		price.String(),
		strconv.FormatInt(receivedAt.Unix(), 10))

	return r
}

func (ct *CLITester) TxMultisigConfirmCall(fromAddress string, callID uint64) *TxRequest {
	r := ct.newTxRequest()
	r.SetCmd(
		"multisig",
		fromAddress,
		"confirm-call",
		strconv.FormatUint(callID, 10))

	return r
}

func (ct *CLITester) QueryCurrenciesIssue(issueID string) (*QueryRequest, *ccTypes.Issue) {
	resObj := &ccTypes.Issue{}
	q := ct.newQueryRequest(resObj)
	q.SetCmd("currencies", "issue", issueID)

	return q, resObj
}

func (ct *CLITester) QueryCurrenciesDestroy(id sdk.Int) (*QueryRequest, *ccTypes.Destroy) {
	resObj := &ccTypes.Destroy{}
	q := ct.newQueryRequest(resObj)
	q.SetCmd("currencies", "destroy", id.String())

	return q, resObj
}

func (ct *CLITester) QueryCurrenciesDestroys(page, limit int) (*QueryRequest, *ccTypes.Destroys) {
	resObj := &ccTypes.Destroys{}
	q := ct.newQueryRequest(resObj)
	q.SetCmd("currencies", "destroys", strconv.Itoa(page), strconv.Itoa(limit))

	return q, resObj
}

func (ct *CLITester) QueryCurrenciesCurrency(symbol string) (*QueryRequest, *ccTypes.Currency) {
	resObj := &ccTypes.Currency{}
	q := ct.newQueryRequest(resObj)
	q.SetCmd("currencies", "currency", symbol)

	return q, resObj
}

func (ct *CLITester) QueryOracleAssets() (*QueryRequest, *oracle.Assets) {
	resObj := &oracle.Assets{}
	q := ct.newQueryRequest(resObj)
	q.SetCmd("oracle", "assets")

	return q, resObj
}

func (ct *CLITester) QueryOracleRawPrices(assetCode string, blockHeight int64) (*QueryRequest, *[]oracle.PostedPrice) {
	resObj := &[]oracle.PostedPrice{}
	q := ct.newQueryRequest(resObj)
	q.SetCmd(
		"oracle",
		"rawprices",
		assetCode,
		strconv.FormatInt(blockHeight, 10))

	return q, resObj
}

func (ct *CLITester) QueryOraclePrice(assetCode string) (*QueryRequest, *oracle.CurrentPrice) {
	resObj := &oracle.CurrentPrice{}
	q := ct.newQueryRequest(resObj)
	q.SetCmd(
		"oracle",
		"price",
		assetCode)

	return q, resObj
}

func (ct *CLITester) QueryMultisigUnique(uniqueID string) (*QueryRequest, *msTypes.CallResp) {
	resObj := &msTypes.CallResp{}
	q := ct.newQueryRequest(resObj)
	q.SetCmd("multisig", "unique", uniqueID)

	return q, resObj
}

func (ct *CLITester) QueryAccount(address string) (*QueryRequest, *auth.BaseAccount) {
	resObj := &auth.BaseAccount{}
	q := ct.newQueryRequest(resObj)
	q.SetCmd("account", address)

	return q, resObj
}

type CLIAccount struct {
	Name            string
	Address         string
	PubKey          string
	Mnemonic        string
	Coins           map[string]sdk.Coin
	IsPOAValidator  bool
	IsOracleNominee bool
}

type CLICmd struct {
	t      *testing.T
	base   string
	args   []string
	inputs string
	proc   *tests.Process
}

func (c *CLICmd) AddArg(flagName, flagValue string) *CLICmd {
	if flagName != "" {
		c.args = append(c.args, fmt.Sprintf("--%s=%s", flagName, flagValue))
	} else {
		c.args = append(c.args, flagValue)
	}

	return c
}

func (c *CLICmd) ChangeArg(oldArg, newArg string) *CLICmd {
	for i := 0; i < len(c.args); i++ {
		if c.args[i] == oldArg {
			c.args[i] = newArg
			break
		}

		if ("--" + c.args[i]) == oldArg {
			c.args[i] = "--" + newArg
			break
		}
	}

	return c
}

func (c *CLICmd) RemoveArg(arg string) *CLICmd {
	for i := 0; i < len(c.args); i++ {
		if c.args[i] == arg || ("--"+c.args[i]) == arg {
			c.args = append(c.args[:i], c.args[i+1:]...)
			break
		}
	}

	return c
}

func (c *CLICmd) String() string {
	return fmt.Sprintf("cmd %q with args [%s] and inputs [%s]", c.base, strings.Join(c.args, " "), c.inputs)
}

func (c *CLICmd) Execute(stdinInput ...string) (retCode int, retStdout, retStderr []byte) {
	c.inputs = strings.Join(stdinInput, ", ")

	proc, err := tests.StartProcess("", c.base, c.args)
	require.NoError(c.t, err, "cmd %q: StartProcess", c.String())

	for _, input := range stdinInput {
		_, err := proc.StdinPipe.Write([]byte(input + "\n"))
		require.NoError(c.t, err, "%s: %q StdinPipe.Write", c.String(), input)
	}

	stdout, stderr, err := proc.ReadAll()
	require.NoError(c.t, err, "%s: reading stdout, stderr", c.String())

	proc.Wait()
	retCode, retStdout, retStderr = proc.ExitState.ExitCode(), stdout, stderr

	return
}

func (c *CLICmd) Start(startLoggers bool) {
	proc, err := tests.CreateProcess("", c.base, c.args)
	require.NoError(c.t, err, "cmd %q: CreateProcess", c.String())

	if startLoggers {
		pipeLogger := func(pipeName string, pipe io.ReadCloser) {
			buf := bufio.NewReader(pipe)
			for {
				line, _, err := buf.ReadLine()
				if err != nil {
					c.t.Logf("%q %s: reading daemon pipe: %v", c.base, pipeName, err)
					c.t.Logf("%q %s: reading daemon pipe: %v", c.base, pipeName, err)
					return
				}

				//ct.t.Logf("%s: %s\n", pipeName, line)
				fmt.Printf("%s: %s\n", pipeName, line)
			}
		}

		go pipeLogger("stdout", proc.StdoutPipe)
		go pipeLogger("stderr", proc.StderrPipe)
	}

	require.NoError(c.t, proc.Cmd.Start(), "cmd %q: Start", c.String())
	c.proc = proc
}

func (c *CLICmd) Stop() {
	require.NotNil(c.t, c.proc, "proc")
	require.NoError(c.t, c.proc.Stop(false), "proc.Stop")
	c.proc = nil
}

func (c *CLICmd) CheckSuccessfulExecute(resultObj interface{}, stdinInput ...string) {
	code, stdout, stderr := c.Execute(stdinInput...)
	require.Equal(c.t, 0, code, "%s: stderr: %s", c.String(), string(stderr))

	if resultObj != nil {
		if err := json.Unmarshal(stdout, resultObj); err == nil {
			return
		}
		if err := json.Unmarshal(stderr, resultObj); err == nil {
			return
		}

		c.t.Fatalf("%s: stdout/stderr unmarshal to object type %t", c.String(), resultObj)
	}
}

type TxRequest struct {
	t              *testing.T
	cdc            *codec.Codec
	cmd            *CLICmd
	accPassphrase  string
	nodeRpcAddress string
}

func (r *TxRequest) SetCmd(module, fromAddress string, args ...string) {
	//cmd.AddArg("broadcast-mode", "block")
	r.cmd.AddArg("", "tx")
	r.cmd.AddArg("", module)

	for _, arg := range args {
		r.cmd.AddArg("", arg)
	}

	if fromAddress != "" {
		r.cmd.AddArg("from", fromAddress)
	}
	r.cmd.AddArg("node", r.nodeRpcAddress)
	r.cmd.AddArg("fees", "1"+config.MainDenom)
	r.cmd.AddArg("", "--yes")
}

func (r *TxRequest) ChangeCmdArg(oldArg, newArg string) {
	r.cmd.ChangeArg(oldArg, newArg)
}

func (r *TxRequest) RemoveCmdArg(arg string) {
	r.cmd.RemoveArg(arg)
}

func (r *TxRequest) Send() (retCode int, retStdout, retStderr []byte) {
	return r.cmd.Execute(r.accPassphrase)
}

func (r *TxRequest) CheckSucceeded() {
	code, stdout, stderr := r.Send()

	require.Equal(r.t, 0, code, "%s: failed with code %d:\nstdout: %s\nstrerr: %s", r.String(), code, string(stdout), string(stderr))
	require.Len(r.t, stderr, 0, "%s: failed with non-empty stderr:\nstdout: %s\nstrerr: %s", r.String(), string(stdout), string(stderr))

	if len(stdout) > 0 {
		txResponse := sdk.TxResponse{}
		require.NoError(r.t, r.cdc.UnmarshalJSON(stdout, &txResponse), "%s: unmarshal", r.String())
		require.Equal(r.t, "", txResponse.Codespace, "%s: SDK codespace", r.String())
		require.Equal(r.t, uint32(0), txResponse.Code, "%s: SDK code", r.String())
	}
}

func (r *TxRequest) CheckFailedWithSDKError(sdkErr sdk.Error) {
	code, stdout, stderr := r.Send()
	require.NotEqual(r.t, 0, code, "%s: succeeded", r.String())
	stdout, stderr = trimCliOutput(stdout), trimCliOutput(stderr)

	txResponse := sdk.TxResponse{}
	stdoutErr := r.cdc.UnmarshalJSON(stdout, &txResponse)
	stderrErr := r.cdc.UnmarshalJSON(stderr, &txResponse)
	if stdoutErr != nil && stderrErr != nil {
		r.t.Fatalf("%s: unmarshal stdout/stderr: %s / %s", r.String(), string(stdout), string(stderr))
	}

	require.Equal(r.t, sdkErr.Codespace(), sdk.CodespaceType(txResponse.Codespace), "%s: codespace", r.String())
	require.Equal(r.t, sdkErr.Code(), sdk.CodeType(txResponse.Code), "%s: code", r.String())
}

func (r *TxRequest) CheckFailedWithErrorSubstring(subStr string) (output string) {
	code, stdout, stderr := r.Send()
	require.NotEqual(r.t, 0, code, "%s: succeeded", r.String())

	stdoutStr, stderrErr := string(stdout), string(stderr)
	output = fmt.Sprintf("stdout: %s\nstderr: %s", stdoutStr, stderrErr)

	if subStr == "" {
		return
	}

	if strings.Contains(stdoutStr, subStr) || strings.Contains(stderrErr, subStr) {
		return
	}
	r.t.Fatalf("%s: stdout/stderr doesn't contain %q sub string", r.String(), subStr)

	return
}

func (r *TxRequest) String() string {
	return fmt.Sprintf("tx %s", r.cmd.String())
}

type QueryRequest struct {
	t              *testing.T
	cdc            *codec.Codec
	cmd            *CLICmd
	nodeRpcAddress string
	resultObj      interface{}
}

func (q *QueryRequest) ChangeCmdArg(oldArg, newArg string) {
	q.cmd.ChangeArg(oldArg, newArg)
}

func (q *QueryRequest) RemoveCmdArg(arg string) {
	q.cmd.RemoveArg(arg)
}

func (q *QueryRequest) SetCmd(module string, args ...string) {
	q.cmd.AddArg("", "query")
	q.cmd.AddArg("", module)

	for _, arg := range args {
		q.cmd.AddArg("", arg)
	}

	q.cmd.AddArg("node", q.nodeRpcAddress)
}

func (q *QueryRequest) CheckSucceeded() {
	code, stdout, stderr := q.cmd.Execute()

	require.Equal(q.t, 0, code, "%s: failed with code %d:\nstdout: %s\nstrerr: %s", q.String(), code, string(stdout), string(stderr))
	require.Len(q.t, stderr, 0, "%s: failed with non-empty stderr:\nstdout: %s\nstrerr: %s", q.String(), string(stdout), string(stderr))

	if q.resultObj != nil {
		err := q.cdc.UnmarshalJSON(stdout, q.resultObj)
		require.NoError(q.t, err, "%s: unmarshal query stdout: %s", q.String(), string(stdout))
	}
}

func (q *QueryRequest) CheckFailedWithSDKError(sdkErr sdk.Error) {
	code, stdout, stderr := q.cmd.Execute()
	require.NotEqual(q.t, 0, code, "%s: succeeded", q.String())
	stdout, stderr = trimCliOutput(stdout), trimCliOutput(stderr)

	qResponse := struct {
		Codespace sdk.CodespaceType `json:"codespace"`
		Code      sdk.CodeType      `json:"code"`
	}{sdk.CodespaceType(""), sdk.CodeType(0)}
	stdoutErr := q.cdc.UnmarshalJSON(stdout, &qResponse)
	stderrErr := q.cdc.UnmarshalJSON(stderr, &qResponse)
	if stdoutErr != nil && stderrErr != nil {
		q.t.Fatalf("%s: unmarshal stdout/stderr: %s / %s", q.String(), string(stdout), string(stderr))
	}

	require.Equal(q.t, sdkErr.Codespace(), qResponse.Codespace, "%s: codespace", q.String())
	require.Equal(q.t, sdkErr.Code(), qResponse.Code, "%s: code", q.String())
}

func (q *QueryRequest) CheckFailedWithErrorSubstring(subStr string) (output string) {
	code, stdout, stderr := q.cmd.Execute()
	require.NotEqual(q.t, 0, code, "%s: succeeded", q.String())

	stdoutStr, stderrErr := string(stdout), string(stderr)
	output = fmt.Sprintf("stdout: %s\nstderr: %s", stdoutStr, stderrErr)

	if subStr == "" {
		return
	}

	if strings.Contains(stdoutStr, subStr) || strings.Contains(stderrErr, subStr) {
		return
	}
	q.t.Fatalf("%s: stdout/stderr doesn't contain %q sub string", q.String(), subStr)

	return
}

func (q *QueryRequest) String() string {
	return fmt.Sprintf("query %s", q.cmd.String())
}

func trimCliOutput(output []byte) []byte {
	for i := 0; i < len(output); i++ {
		if output[i] == '{' {
			output = output[i:]
			break
		}
	}

	return output
}

func RequireCliTestEnv(t *testing.T) {
	const envName = "DN_CLI_TESTS"

	if os.Getenv(envName) == "" {
		t.Skipf("skipping test: %q not set", envName)
	}
}
