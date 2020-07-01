package clitester

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/99designs/keyring"
	"github.com/cosmos/cosmos-sdk/codec"
	sdkKeys "github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govTypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/require"
	tmCTypes "github.com/tendermint/tendermint/rpc/core/types"
	tmRPCTypes "github.com/tendermint/tendermint/rpc/lib/types"
	tmTypes "github.com/tendermint/tendermint/types"

	dnConfig "github.com/dfinance/dnode/cmd/config"
	"github.com/dfinance/dnode/helpers/tests/utils"
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
	GovernanceConfig  GovernanceConfig
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
		GovernanceConfig:  NewGovernanceConfig(),
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

// GenesisState reads genesisState file and returns unmarshalled result.
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

// RestartDaemon restarts dnode daemon process.
func (ct *CLITester) RestartDaemon(waitForStart, printLogs bool) {
	require.NotNil(ct.t, ct.daemon, "daemon is not running")

	ct.daemon.Stop()
	ct.daemon = nil

	ct.startDemon(waitForStart, printLogs)
}

// CheckDaemonStopped waits for dnode daemon to stop.
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

// DaemonLogsContain checks dnode daemon logs contains some strings.
func (ct *CLITester) DaemonLogsContain(subStr string) bool {
	require.NotNil(ct.t, ct.daemon, "daemon wasn't started")

	return ct.daemon.LogsContain(subStr)
}

// StartRestServer start REST server via dncli cmd.
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

// SetVMCompilerAddressNet sets DVM address using TCP connection.
//   <skipTcpTest> flag omits TCP port ping tests.
func (ct *CLITester) SetVMCompilerAddressNet(address string, skipTcpTest bool) {
	ct.VMConnection.CompilerAddress = address
	if !skipTcpTest {
		require.NoError(ct.t, utils.PingTcpAddress(address, 500*time.Millisecond), "VM compiler address (net)")
	}
}

// SetVMCompilerAddressUDS sets DVM address using UDS connection.
func (ct *CLITester) SetVMCompilerAddressUDS(path string) {
	_, err := os.Stat(path)
	require.NoError(ct.t, err, "VM compiler address (UDS)")
	ct.VMConnection.CompilerAddress = "unix://" + path
}

// UpdateAccountBalance updates ct.Accounts balance for specified name.
// Contract: works only for non-module accounts.
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

// UpdateAccountsBalance updates all ct.Accounts balances.
func (ct *CLITester) UpdateAccountsBalance() {
	for accName, prevAcc := range ct.Accounts {
		var curCoins sdk.Coins
		var accNumber uint64

		if prevAcc.IsModuleAcc {
			q, modAcc := ct.QueryModuleAccount(prevAcc.Address)
			q.CheckSucceeded()

			curCoins, accNumber = modAcc.Coins, modAcc.AccountNumber
		} else {
			q, baseAcc := ct.QueryAccount(prevAcc.Address)
			q.CheckSucceeded()

			curCoins, accNumber = baseAcc.Coins, baseAcc.AccountNumber
		}

		prevAcc.Number = accNumber
		for _, curCoin := range curCoins {
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

// UpdateAccountsBalance returns current blockHeight using RPC methods.
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

// WaitForNextBlocks waits for next <n> blocks and return current blockHeight.
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

// ConfirmCall confirms multisig call with validator accounts.
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

// SubmitAndConfirmProposal submits proposal, votes for it and waits until Passed / Rejected.
// Contract 1: plannedBlockHeight must be -1 on txRequest creation.
// Contract 2: proposalTx must cover proposal minDeposit.
// Contract 3: not concurrent safe.
func (ct *CLITester) SubmitAndConfirmProposal(proposalTx *TxRequest, isPlannedProposal bool) {
	// get current proposals count
	prevProposalsCount := 0
	{
		q, proposals := ct.QueryGovProposals(-1, -1, nil, nil, nil)
		out, err := q.Execute()
		if err == nil {
			prevProposalsCount = len(*proposals)
		} else {
			if !strings.Contains(out, "no matching proposals found") {
				require.NoError(ct.t, err, "initial proposal query failed")
			}
		}
	}

	plannedHeight := int64(0)
	if isPlannedProposal {
		// calculate the planned height
		curHeight, err := ct.GetCurrentBlockHeight()
		require.NoError(ct.t, err, "GetCurrentBlockHeight failed")
		plannedHeight = curHeight + 20

		// modify the plannedBlockHeight argument
		proposalTx.ChangeCmdArg("-1", strconv.FormatInt(plannedHeight, 10))
	}

	// emit the Tx
	proposalTx.CheckSucceeded()

	// check proposal added
	proposalID := uint64(0)
	{
		q, proposals := ct.QueryGovProposals(-1, -1, nil, nil, nil)
		q.CheckSucceeded()

		curProposalsCount := len(*proposals)
		require.Equal(ct.t, curProposalsCount-1, prevProposalsCount, "proposal not added")

		proposal := (*proposals)[curProposalsCount-1]
		require.Equal(ct.t, proposal.Status, govTypes.StatusVotingPeriod, "invalid proposal initial state")
		proposalID = proposal.ProposalID
	}

	// vote (CLITester starts one node, so we need to add only one vote)
	ct.TxGovVote(ct.Accounts["pos"].Address, proposalID, govTypes.OptionYes).CheckSucceeded()

	// wait for voting period
	time.Sleep(ct.GovernanceConfig.MinVotingDur)

	// check proposal passed
	{
		q, proposal := ct.QueryGovProposal(proposalID)
		q.CheckSucceeded()

		require.Equal(ct.t, proposal.Status, gov.StatusPassed, "proposal didn't pass")
	}

	if isPlannedProposal {
		// wait for scheduler
		{
			curHeight, err := ct.GetCurrentBlockHeight()
			require.NoError(ct.t, err, "GetCurrentBlockHeight failed")

			if curHeight < plannedHeight {
				ct.WaitForNextBlocks(plannedHeight - curHeight)
			}
		}
	}
}

// CreateAccount creates a new account with specified balances.
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

// Close stop dnode tester.
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
