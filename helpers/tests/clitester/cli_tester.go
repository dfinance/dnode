package clitester

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdkKeys "github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	tmCTypes "github.com/tendermint/tendermint/rpc/core/types"
	tmRPCTypes "github.com/tendermint/tendermint/rpc/lib/types"
	tmTypes "github.com/tendermint/tendermint/types"

	"github.com/dfinance/dnode/cmd/config"
	dnConfig "github.com/dfinance/dnode/cmd/config"
	"github.com/dfinance/dnode/helpers/tests"
)

type CLITester struct {
	RootDir           string
	ChainID           string
	MonikerID         string
	AccountPassphrase string
	Accounts          map[string]*CLIAccount
	Cdc               *codec.Codec
	VmListenPort      string
	t                 *testing.T
	validatorAddrs    []string
	wbdBinary         string
	wbcliBinary       string
	rpcAddress        string
	rpcPort           string
	p2pAddress        string
	vmConnectAddress  string
	vmListenAddress   string
	vmCompilerAddress string
	daemon            *CLICmd
}

type CLIAccount struct {
	Name            string
	Address         string
	EthAddress      string
	PubKey          string
	Mnemonic        string
	Coins           map[string]sdk.Coin
	IsPOAValidator  bool
	IsOracleNominee bool
}

func New(t *testing.T, printDaemonLogs bool) *CLITester {
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

	ct := CLITester{
		t:                 t,
		Cdc:               makeCodec(),
		wbdBinary:         "dnode",
		wbcliBinary:       "dncli",
		ChainID:           "test-chain",
		MonikerID:         "test-moniker",
		AccountPassphrase: "passphrase",
		rpcAddress:        srvAddr,
		rpcPort:           srvPort,
		p2pAddress:        p2pAddr,
		vmConnectAddress:  fmt.Sprintf("127.0.0.1:%s", vmConnectPort),
		VmListenPort:      vmListenPort,
		vmListenAddress:   fmt.Sprintf("127.0.0.1:%s", vmListenPort),
		vmCompilerAddress: "",
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
	ct.Accounts["plain"] = &CLIAccount{
		Coins: map[string]sdk.Coin{
			config.MainDenom: sdk.NewCoin(config.MainDenom, smallAmount),
		},
	}

	rootDir, err := ioutil.TempDir("/tmp", fmt.Sprintf("wd-cli-test-%s-", ct.t.Name()))
	require.NoError(t, err, "TempDir")
	ct.RootDir = rootDir

	ct.initChain()

	ct.startDemon(true, printDaemonLogs)

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
	ethAddresses := []string{
		"0x82A978B3f5962A5b0957d9ee9eEf472EE55B42F1",
		"0x7d577a597B2742b498Cb5Cf0C26cDCD726d39E6e",
		"0xDCEceAF3fc5C0a63d195d69b1A90011B7B19650D",
		"0x598443F1880Ef585B21f1d7585Bd0577402861E5",
		"0x13cBB8D99C6C4e0f2728C7d72606e78A29C4E224",
		"0x77dB2BEBBA79Db42a978F896968f4afCE746ea1F",
		"0x24143873e0E0815fdCBcfFDbe09C979CbF9Ad013",
		"0x10A1c1CB95c92EC31D3f22C66Eef1d9f3F258c6B",
		"0xe0FC04FA2d34a66B779fd5CEe748268032a146c0",
	}

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
				accValue.EthAddress = ethAddresses[poaValidatorIdx]

				cmd := ct.newWbdCmd().
					AddArg("", "add-genesis-poa-validator").
					AddArg("", accValue.Address).
					AddArg("", accValue.EthAddress)
				poaValidatorIdx++

				cmd.CheckSuccessfulExecute(nil)
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

	// VM default write sets
	{
		defWriteSetsPath := "${GOPATH}/src/github.com/dfinance/dnode/x/vm/internal/keeper/genesis_ws.json"
		defWriteSetsPath = os.ExpandEnv(defWriteSetsPath)

		cmd := ct.newWbdCmd().
			AddArg("", "read-genesis-write-set").
			AddArg("", defWriteSetsPath).
			AddArg("home", ct.RootDir)

		cmd.CheckSuccessfulExecute(nil, ct.AccountPassphrase)
	}

	// change default genesis params
	{
		appState := ct.GenesisState()

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

func (ct *CLITester) startDemon(waitForStart, printLogs bool) {
	const startRetries = 100

	require.Nil(ct.t, ct.daemon)

	cmd := ct.newWbdCmd().
		AddArg("", "start").
		AddArg("rpc.laddr", ct.rpcAddress).
		AddArg("p2p.laddr", ct.p2pAddress)
	cmd.Start(ct.t, printLogs)

	// wait for the node to start up
	if waitForStart {
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

	ct.daemon = cmd
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

func (ct *CLITester) GenesisState() GenesisState {
	genesisFile := server.NewDefaultContext().Config.Genesis
	genesisPath := path.Join(ct.RootDir, genesisFile)

	cdc := codec.New()
	genDoc, err := tmTypes.GenesisDocFromFile(genesisPath)
	require.NoError(ct.t, err, "reading genesis file %q", genesisPath)

	appState := GenesisState{}
	require.NoError(ct.t, cdc.UnmarshalJSON(genDoc.AppState, &appState), "unmarshal appState")

	return appState
}

func (ct *CLITester) RestartDaemon(waitForStart, printLogs bool) {
	require.NotNil(ct.t, ct.daemon, "daemon is not running")

	ct.daemon.Stop()
	ct.daemon = nil

	ct.startDemon(waitForStart, printLogs)
}

func (ct *CLITester) CheckDaemonStopped(timeout time.Duration) (exitCode int, daemonLogs []string) {
	require.NotNil(ct.t, ct.daemon, "daemon wasn't started")

	retCode := ct.daemon.WaitForStop(timeout)
	require.NotNil(ct.t, retCode, "daemon didn't stop in %v", timeout)

	exitCode = *retCode
	daemonLogs = make([]string, len(ct.daemon.logs))
	copy(daemonLogs, ct.daemon.logs)

	ct.daemon = nil

	return
}

func (ct *CLITester) DaemonLogsContain(subStr string) bool {
	require.NotNil(ct.t, ct.daemon, "daemon wasn't started")

	return ct.daemon.LogsContain(subStr)
}

func (ct *CLITester) SetVMCompilerAddress(address string) {
	ct.vmCompilerAddress = address
	require.NoError(ct.t, tests.PingTcpAddress(address), "VM compiler address")
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
		if curHeight-prevHeight >= n {
			return curHeight
		}
	}
}

func (ct *CLITester) ConfirmCall(uniqueID string) {
	// get validators list
	validatorAddrs := make([]string, 0)
	for _, acc := range ct.Accounts {
		if acc.IsPOAValidator {
			validatorAddrs = append(validatorAddrs, acc.Address)
		}
	}
	requiredVotes := len(validatorAddrs)/2 + 1

	// get callID
	q, call := ct.QueryMultiSigUnique(uniqueID)
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
	availableVotes := len(call.Votes) + len(validatorAddrs)
	require.LessOrEqual(ct.t, requiredVotes, availableVotes, "not enough validators to confirm call")

	// send confirms
	for i := 0; i < requiredVotes-len(call.Votes); i++ {
		ct.TxMultiSigConfirmCall(validatorAddrs[i], call.Call.MsgID).CheckSucceeded()
	}
	ct.WaitForNextBlocks(2)

	// check call approved
	{
		q, call := ct.QueryMultiSigUnique(uniqueID)
		q.CheckSucceeded()
		require.Equal(ct.t, uniqueID, call.Call.UniqueID)
		require.True(ct.t, call.Call.Approved)
	}
}
