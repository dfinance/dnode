package clitester

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"sync"
	"testing"
	"time"

	"github.com/99designs/keyring"
	"github.com/cosmos/cosmos-sdk/codec"
	sdkKeys "github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	tmCTypes "github.com/tendermint/tendermint/rpc/core/types"
	tmRPCTypes "github.com/tendermint/tendermint/rpc/lib/types"
	tmTypes "github.com/tendermint/tendermint/types"

	dnConfig "github.com/dfinance/dnode/cmd/config"
	"github.com/dfinance/dnode/helpers/tests"
)

var (
	cfgMtx sync.Mutex
)

type CLITester struct {
	Cdc               *codec.Codec
	IDs               NodeIdConfig
	Dirs              DirConfig
	BinaryPath        BinaryPathConfig
	Accounts          map[string]*CLIAccount
	Currencies        map[string]CurrencyInfo
	AccountPassphrase string
	DefAssetCode      string
	NodePorts         NodePortConfig
	VMConnection      VMConnectionConfig
	VMCommunication   VMCommunicationConfig
	ConsensusTimings  ConsensusTimingConfig
	//
	t *testing.T
	//
	keyBase        sdkKeys.Keybase
	keyringBackend keyring.BackendType
	//
	restServer  *CLICmd
	restAddress string
	//
	daemonLogLvl string
	daemon       *CLICmd
}

func New(t *testing.T, printDaemonLogs bool, options ...CLITesterOption) *CLITester {
	sdkConfig := sdk.GetConfig()
	dnConfig.InitBechPrefixes(sdkConfig)

	ct := CLITester{
		IDs:               NewTestNodeIdConfig(),
		BinaryPath:        NewTestBinaryPathConfig(),
		Currencies:        NewCurrencyMap(),
		VMCommunication:   NewTestVMCommunicationConfig(),
		ConsensusTimings:  NewTestConsensusTimingConfig(),
		AccountPassphrase: "passphrase",
		DefAssetCode:      "tst_tst",
		//
		t:   t,
		Cdc: makeCodec(),
		//
		keyBase:        sdkKeys.NewInMemory(),
		keyringBackend: keyring.FileBackend,
	}

	dirs, err := NewTempDirConfig(ct.t.Name())
	require.NoError(ct.t, err, "New: dirs")
	ct.Dirs = dirs

	accounts, err := NewAccountMap()
	require.NoError(ct.t, err, "New: accounts")
	ct.Accounts = accounts

	nodePorts, err := NewTestNodePortConfig()
	require.NoError(ct.t, err, "New: node ports")
	ct.NodePorts = nodePorts

	vmConnection, err := NewTestVMConnectionConfigTCP()
	require.NoError(ct.t, err, "New: VM connection via TCP")
	ct.VMConnection = vmConnection

	for _, option := range options {
		require.NoError(ct.t, option(&ct), "option failed")
	}

	ct.initChain()

	ct.startDemon(true, printDaemonLogs)

	ct.UpdateAccountsBalance()

	http.DefaultTransport.(*http.Transport).MaxIdleConnsPerHost = 100

	return &ct
}

func (ct *CLITester) newWbdCmd() *CLICmd {
	cmd := &CLICmd{t: ct.t, base: ct.BinaryPath.wbd}
	cmd.AddArg("home", ct.Dirs.RootDir)

	return cmd
}

func (ct *CLITester) newWbcliCmd() *CLICmd {
	cmd := &CLICmd{t: ct.t, base: ct.BinaryPath.wbcli}
	cmd.AddArg("home", ct.Dirs.DncliDir)
	cmd.AddArg("chain-id", ct.IDs.ChainID)
	cmd.AddArg("output", "json")

	return cmd
}

func (ct *CLITester) newRestRequest() *RestRequest {
	return &RestRequest{
		t:       ct.t,
		cdc:     ct.Cdc,
		baseUrl: ct.restAddress,
		gas:     DefaultGas,
	}
}

func (ct *CLITester) newTxRequest() *TxRequest {
	return &TxRequest{
		t:              ct.t,
		cdc:            ct.Cdc,
		cmd:            ct.newWbcliCmd(),
		nodeRpcAddress: ct.NodePorts.RPCAddress,
		accPassphrase:  ct.AccountPassphrase,
		gas:            DefaultGas,
	}
}

func (ct *CLITester) newQueryRequest(resultObj interface{}) *QueryRequest {
	return &QueryRequest{
		t:              ct.t,
		cdc:            ct.Cdc,
		cmd:            ct.newWbcliCmd(),
		nodeRpcAddress: ct.NodePorts.RPCAddress,
		resultObj:      resultObj,
	}
}

func (ct *CLITester) startDemon(waitForStart, printLogs bool) {
	const startTimeout = 30 * time.Second

	require.Nil(ct.t, ct.daemon)

	cmd := ct.newWbdCmd().
		AddArg("", "start").
		AddArg("rpc.laddr", ct.NodePorts.RPCAddress).
		AddArg("p2p.laddr", ct.NodePorts.P2PAddress)
	if ct.daemonLogLvl != "" {
		cmd.AddArg("log_level", ct.daemonLogLvl)
	}
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
	genesisPath := path.Join(ct.Dirs.RootDir, genesisFile)

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
	genesisPath := path.Join(ct.Dirs.RootDir, genesisFile)

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
	cmd := &CLICmd{t: ct.t, base: ct.BinaryPath.wbcli}
	cmd.AddArg("", "rest-server")
	cmd.AddArg("laddr", "tcp://"+restAddress)
	cmd.AddArg("node", ct.NodePorts.RPCAddress)
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

func (ct *CLITester) SetVMCompilerAddressNet(address string, skipTcpTest bool) {
	ct.VMConnection.CompilerAddress = address
	if !skipTcpTest {
		require.NoError(ct.t, tests.PingTcpAddress(address, 500*time.Millisecond), "VM compiler address (net)")
	}
}

func (ct *CLITester) SetVMCompilerAddressUDS(path string) {
	_, err := os.Stat(path)
	require.NoError(ct.t, err, "VM compiler address (UDS)")
	ct.VMConnection.CompilerAddress = "unix://" + path
}

func (ct *CLITester) UpdateAccountBalance(name string) {
	account, ok := ct.Accounts[name]
	require.True(ct.t, ok, "account %q: not found", name)

	q, acc := ct.QueryAccount(account.Address)
	q.CheckSucceeded()

	account.Number = acc.AccountNumber
	for _, coin := range acc.Coins {
		account.Coins[coin.Denom] = coin
	}
}

func (ct *CLITester) UpdateAccountsBalance() {
	for accName, prevAcc := range ct.Accounts {
		q, curAcc := ct.QueryAccount(prevAcc.Address)
		q.CheckSucceeded()

		prevAcc.Number = curAcc.AccountNumber
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
	url := fmt.Sprintf("http://localhost:%s/block", ct.NodePorts.RPCPort)

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

func (ct *CLITester) CreateAccount(name string, balances ...StringPair) {
	account := &CLIAccount{Coins: make(map[string]sdk.Coin)}

	// create key
	{
		cmd := ct.newWbcliCmd().
			AddArg("", "keys").
			AddArg("", "add").
			AddArg("", name)
		output := sdkKeys.KeyOutput{}
		cmd.CheckSuccessfulExecute(&output, ct.AccountPassphrase, ct.AccountPassphrase)

		account.Name = name
		account.Address = output.Address
		account.PubKey = output.PubKey
		account.Mnemonic = output.Mnemonic
	}

	// add key to CT keychain
	{
		cmd := ct.newWbcliCmd().
			AddArg("", "keys").
			AddArg("", "export").
			AddArg("", name)

		output := cmd.CheckSuccessfulExecute(nil, ct.AccountPassphrase, ct.AccountPassphrase, ct.AccountPassphrase, ct.AccountPassphrase)
		require.NoError(ct.t, ct.keyBase.ImportPrivKey(name, output, ct.AccountPassphrase), "account %q: keyBase.ImportPrivKey", name)
	}

	// issue coins
	for _, balance := range balances {
		denom, amountStr := balance.Key, balance.Value
		issueID := fmt.Sprintf("%s_%s", name, denom)

		amount, ok := sdk.NewIntFromString(amountStr)
		require.True(ct.t, ok, "invalid amount: %s", amountStr)

		currency, ok := ct.Currencies[denom]
		require.True(ct.t, ok, "currency with %q denom: not found", denom)

		validator1Address := ct.Accounts["validator1"].Address
		ct.TxCurrenciesIssue(account.Address, validator1Address, denom, amount, currency.Decimals, issueID).CheckSucceeded()

		ct.ConfirmCall(issueID)
	}

	// store and update balances
	ct.Accounts[name] = account
	ct.UpdateAccountBalance(name)
}

func (ct *CLITester) Close() {
	if ct.restServer != nil {
		ct.restServer.Stop()
	}

	if ct.daemon != nil {
		ct.daemon.Stop()
	}

	if ct.Dirs.RootDir != "" {
		os.RemoveAll(ct.Dirs.RootDir)
	}
}
