package clitester

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/99designs/keyring"
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

var (
	cfgMtx sync.Mutex
)

type CLITester struct {
	RootDir  string
	DncliDir string
	UDSDir   string
	//
	ChainID   string
	MonikerID string
	//
	AccountPassphrase string
	Accounts          map[string]*CLIAccount
	//
	Cdc          *codec.Codec
	DefAssetCode string
	//
	VmListenPort  string
	VmConnectPort string
	//
	//
	t           *testing.T
	keyBase     sdkKeys.Keybase
	wbdBinary   string
	wbcliBinary string
	//
	rpcAddress string
	rpcPort    string
	p2pAddress string
	//
	vmBaseAddress     string
	vmConnectAddress  string
	vmListenAddress   string
	vmCompilerAddress string
	vmComMinBackoffMs int
	vmComMaxBackoffMs int
	vmComMaxAttempts  int
	//
	restAddress    string
	daemon         *CLICmd
	restServer     *CLICmd
	keyringBackend keyring.BackendType
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
	IsOracle        bool
}

func New(t *testing.T, printDaemonLogs bool, options ...CLITesterOption) *CLITester {
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
		keyBase:           sdkKeys.NewInMemory(),
		ChainID:           "test-chain",
		MonikerID:         "test-moniker",
		AccountPassphrase: "passphrase",
		DefAssetCode:      "tst",
		keyringBackend:    keyring.FileBackend,
		//
		rpcAddress: srvAddr,
		rpcPort:    srvPort,
		p2pAddress: p2pAddr,
		//
		VmConnectPort:     vmConnectPort,
		VmListenPort:      vmListenPort,
		vmBaseAddress:     "127.0.0.1",
		vmComMinBackoffMs: 100,
		vmComMaxBackoffMs: 150,
		vmComMaxAttempts:  1,
		vmCompilerAddress: "",
		//
		Accounts: make(map[string]*CLIAccount, 0),
	}
	ct.vmConnectAddress = fmt.Sprintf("%s:%s", ct.vmBaseAddress, ct.VmConnectPort)
	ct.vmListenAddress = fmt.Sprintf("%s:%s", ct.vmBaseAddress, ct.VmListenPort)

	smallAmount, ok := sdk.NewIntFromString("1000000000000000000000")
	require.True(t, ok, "NewInt for smallAmount")
	bigAmount, ok := sdk.NewIntFromString("1000000000000000000000")
	//bigAmount, ok := sdk.NewIntFromString("90000000000000000000000000")
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
	ct.Accounts["validator4"] = &CLIAccount{
		Coins: map[string]sdk.Coin{
			config.MainDenom: sdk.NewCoin(config.MainDenom, smallAmount),
		},
		IsPOAValidator: true,
	}
	ct.Accounts["validator5"] = &CLIAccount{
		Coins: map[string]sdk.Coin{
			config.MainDenom: sdk.NewCoin(config.MainDenom, smallAmount),
		},
		IsPOAValidator: true,
	}
	ct.Accounts["nominee"] = &CLIAccount{
		Coins: map[string]sdk.Coin{
			config.MainDenom: sdk.NewCoin(config.MainDenom, smallAmount),
		},
		IsOracleNominee: true,
	}
	ct.Accounts["oracle1"] = &CLIAccount{
		Coins: map[string]sdk.Coin{
			config.MainDenom: sdk.NewCoin(config.MainDenom, smallAmount),
		},
		IsOracle: true,
	}
	ct.Accounts["oracle2"] = &CLIAccount{
		Coins: map[string]sdk.Coin{
			config.MainDenom: sdk.NewCoin(config.MainDenom, smallAmount),
		},
		IsOracle: false,
	}
	ct.Accounts["oracle3"] = &CLIAccount{
		Coins: map[string]sdk.Coin{
			config.MainDenom: sdk.NewCoin(config.MainDenom, smallAmount),
		},
		IsOracle: false,
	}
	ct.Accounts["plain"] = &CLIAccount{
		Coins: map[string]sdk.Coin{
			config.MainDenom: sdk.NewCoin(config.MainDenom, smallAmount),
		},
	}

	rootDir, err := ioutil.TempDir("/tmp", fmt.Sprintf("wd-cli-test-%s-", ct.t.Name()))
	require.NoError(t, err, "TempDir")
	ct.RootDir = rootDir
	ct.DncliDir = path.Join(rootDir, "dncli")

	ct.UDSDir = path.Join(ct.RootDir, "sockets")
	require.NoError(t, os.Mkdir(ct.UDSDir, 0777), "creating sockets dir")

	for _, option := range options {
		require.NoError(ct.t, option(&ct), "option failed")
	}

	ct.initChain()

	ct.startDemon(true, printDaemonLogs)

	ct.UpdateAccountsBalance()

	return &ct
}

func (ct *CLITester) Close() {
	if ct.restServer != nil {
		ct.restServer.Stop()
	}

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
	cmd.AddArg("home", ct.DncliDir)
	cmd.AddArg("chain-id", ct.ChainID)
	cmd.AddArg("output", "json")

	return cmd
}

func (ct *CLITester) newRestRequest() *RestRequest {
	return &RestRequest{
		t:       ct.t,
		cdc:     ct.Cdc,
		baseUrl: ct.restAddress,
	}
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

	// configure dncli
	{
		cmd := ct.newWbcliCmd().
			AddArg("", "config").
			AddArg("", "keyring-backend").
			AddArg("", string(ct.keyringBackend))
		cmd.CheckSuccessfulExecute(nil)
	}

	// adjust tendermint config (make blocks generation faster)
	{
		cfgMtx.Lock()
		defer cfgMtx.Unlock()

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

			// get armored private key
			{
				cmd := ct.newWbcliCmd().
					AddArg("", "keys").
					AddArg("", "export").
					AddArg("", accName)

				output := cmd.CheckSuccessfulExecute(nil, ct.AccountPassphrase, ct.AccountPassphrase, ct.AccountPassphrase, ct.AccountPassphrase)
				require.NoError(ct.t, ct.keyBase.ImportPrivKey(accName, output, ct.AccountPassphrase), "account %q: keyBase.ImportPrivKey", accName)
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

				cmd.CheckSuccessfulExecute(nil, ct.AccountPassphrase)
			}

			// POA validator
			if accValue.IsPOAValidator {
				require.True(ct.t, poaValidatorIdx < len(EthAddresses), "add more predefined ethAddresses")
				accValue.EthAddress = EthAddresses[poaValidatorIdx]

				cmd := ct.newWbdCmd().
					AddArg("", "add-genesis-poa-validator").
					AddArg("", accValue.Address).
					AddArg("", accValue.EthAddress)
				cmd.CheckSuccessfulExecute(nil, ct.AccountPassphrase)

				poaValidatorIdx++
			}

			// Oracle nominee
			if accValue.IsOracleNominee {
				cmd := ct.newWbdCmd().
					AddArg("", "add-oracle-nominees-gen").
					AddArg("", accValue.Address)

				cmd.CheckSuccessfulExecute(nil, ct.AccountPassphrase)
			}
		}
	}

	// validator genTX
	{
		cmd := ct.newWbdCmd().
			AddArg("", "gentx").
			AddArg("home-client", ct.DncliDir).
			AddArg("name", "pos").
			AddArg("amount", ct.Accounts["pos"].Coins[config.MainDenom].String()).
			AddArg("keyring-backend", string(ct.keyringBackend))

		cmd.CheckSuccessfulExecute(nil, ct.AccountPassphrase, ct.AccountPassphrase, ct.AccountPassphrase)
	}

	// VM default write sets
	{
		defWriteSetsPath := os.ExpandEnv(DefVmWriteSetsPath)

		cmd := ct.newWbdCmd().
			AddArg("", "read-genesis-write-set").
			AddArg("", defWriteSetsPath).
			AddArg("home", ct.RootDir)

		cmd.CheckSuccessfulExecute(nil)
	}

	// add Oracle assets
	{
		oracles := make([]string, 0)
		oracles = append(oracles, ct.Accounts["oracle1"].Address)
		oracles = append(oracles, ct.Accounts["oracle2"].Address)

		cmd := ct.newWbdCmd().
			AddArg("", "add-oracle-asset-gen").
			AddArg("", ct.DefAssetCode).
			AddArg("", strings.Join(oracles, ","))

		cmd.CheckSuccessfulExecute(nil)
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
		vmConfig.InitialBackoff = ct.vmComMinBackoffMs
		vmConfig.MaxBackoff = ct.vmComMaxBackoffMs
		vmConfig.MaxAttempts = ct.vmComMaxAttempts
		dnConfig.WriteVMConfig(ct.RootDir, vmConfig)
	}
}

func (ct *CLITester) startDemon(waitForStart, printLogs bool) {
	const startTimeout = 30 * time.Second

	require.Nil(ct.t, ct.daemon)

	cmd := ct.newWbdCmd().
		AddArg("", "start").
		AddArg("rpc.laddr", ct.rpcAddress).
		AddArg("p2p.laddr", ct.p2pAddress)
	cmd.Start(ct.t, printLogs)

	// wait for the node to start up
	if waitForStart {
		timeoutCh := time.NewTimer(startTimeout).C
		for {
			blockHeight, err := ct.GetCurrentBlockHeight()
			if err == nil && blockHeight > 1 {
				break
			}

			select {
			case <-timeoutCh:
				ct.t.Fatalf("wait for the node to start up (%v): failed", startTimeout)
			default:
				time.Sleep(50 * time.Millisecond)
			}
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

func (ct *CLITester) StartRestServer(printLogs bool) (restUrl string) {
	const startTimeout = 30 * time.Second

	require.Nil(ct.t, ct.restServer)

	_, restPort, err := server.FreeTCPAddr()
	require.NoError(ct.t, err, "FreeTCPAddr for REST server")
	restAddress := "localhost:" + restPort
	restUrl = "http://" + restAddress

	//cmd := ct.newWbcliCmd().
	cmd := &CLICmd{t: ct.t, base: ct.wbcliBinary}
	cmd.AddArg("", "rest-server")
	cmd.AddArg("laddr", "tcp://"+restAddress)
	cmd.AddArg("node", ct.rpcAddress)
	cmd.AddArg("trust-node", "true")
	cmd.Start(ct.t, printLogs)

	// wait for the server to start up
	timeoutCh := time.NewTimer(startTimeout).C
	for {
		resp, err := http.Get(restUrl + "/node_info")
		if err == nil && resp.StatusCode == http.StatusOK {
			break
		}

		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}

		select {
		case <-timeoutCh:
			ct.t.Fatalf("wait for the REST server to start (%v): failed", startTimeout)
		default:
			time.Sleep(50 * time.Millisecond)
		}
	}

	ct.restServer, ct.restAddress = cmd, restUrl

	return
}

func (ct *CLITester) SetVMCompilerAddressNet(address string) {
	ct.vmCompilerAddress = address
	require.NoError(ct.t, tests.PingTcpAddress(address), "VM compiler address (net)")
}

func (ct *CLITester) SetVMCompilerAddressUDS(path string) {
	_, err := os.Stat(path)
	require.NoError(ct.t, err, "VM compiler address (UDS)")
	ct.vmCompilerAddress = "unix://" + path
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
	ct.WaitForNextBlocks(1)

	// check call approved
	{
		q, call := ct.QueryMultiSigUnique(uniqueID)
		q.CheckSucceeded()
		require.Equal(ct.t, uniqueID, call.Call.UniqueID)
		require.True(ct.t, call.Call.Approved)
	}
}
