package app

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/WingsDao/wings-blockchain/x/core"
	msMsgs "github.com/WingsDao/wings-blockchain/x/multisig/msgs"
	poaTypes "github.com/WingsDao/wings-blockchain/x/poa/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdkKeybase "github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkRest "github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/auth"
	sdkAuthRest "github.com/cosmos/cosmos-sdk/x/auth/client/rest"
	"github.com/cosmos/cosmos-sdk/x/genaccounts"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	tmFlags "github.com/tendermint/tendermint/libs/cli/flags"
	"github.com/tendermint/tendermint/libs/log"
	tmNode "github.com/tendermint/tendermint/node"
	"github.com/tendermint/tendermint/p2p"
	"github.com/tendermint/tendermint/privval"
	"github.com/tendermint/tendermint/proxy"
	rpcClient "github.com/tendermint/tendermint/rpc/client"
	rpcserver "github.com/tendermint/tendermint/rpc/lib/server"
	tmTypes "github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tm-db"
	"google.golang.org/grpc"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"sort"
	"testing"
	"time"

	vmConfig "github.com/WingsDao/wings-blockchain/cmd/config"
	wbConfig "github.com/WingsDao/wings-blockchain/cmd/config"
	msTypes "github.com/WingsDao/wings-blockchain/x/multisig/types"
	"github.com/WingsDao/wings-blockchain/x/vm"

	tmDb "github.com/tendermint/tm-db"
	tmDbm "github.com/tendermint/tm-db"
)

var (
	chainID         = ""
	currency1Symbol = "testcoin1"
	currency2Symbol = "testcoin2"
	currency3Symbol = "testcoin3"
	issue1ID        = "issue1"
	issue2ID        = "issue2"
	issue3ID        = "issue3"
	amount          = sdk.NewInt(100)
	ethAddresses    = []string{
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
)

// copy of SDK's lcd.RestServer, but stoppable (the original has no graceful shutdown) + embedded configuration
// TODO: newer SDK version does have a better lcd.RestServer, remove this on SDK dependency change
type StoppableRestServer struct {
	Mux     *mux.Router
	CliCtx  context.CLIContext
	KeyBase sdkKeybase.Keybase

	log      log.Logger
	listener net.Listener
}

func NewStoppableRestServer(cdc *codec.Codec, customRpcClient rpcClient.Client) *StoppableRestServer {
	r := mux.NewRouter()
	cliCtx := context.NewCLIContext().WithCodec(cdc).WithTrustNode(true).WithClient(customRpcClient)
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "rest-server")

	rs := &StoppableRestServer{
		Mux:    r,
		CliCtx: cliCtx,
		log:    logger,
	}

	client.RegisterRoutes(rs.CliCtx, rs.Mux)
	sdkAuthRest.RegisterTxRoutes(rs.CliCtx, rs.Mux)
	ModuleBasics.RegisterRESTRoutes(rs.CliCtx, rs.Mux)

	return rs
}

func (rs *StoppableRestServer) Start(listenAddr string, maxOpen int, readTimeout, writeTimeout uint) (err error) {
	server.TrapSignal(func() {
		err := rs.listener.Close()
		rs.log.Error("error closing listener", "err", err)
	})

	cfg := rpcserver.DefaultConfig()
	cfg.MaxOpenConnections = maxOpen
	cfg.ReadTimeout = time.Duration(readTimeout) * time.Second
	cfg.WriteTimeout = time.Duration(writeTimeout) * time.Second

	rs.listener, err = rpcserver.Listen(listenAddr, cfg)
	if err != nil {
		return
	}
	rs.log.Info("Starting application REST service...")

	go func() {
		if err := rpcserver.StartHTTPServer(rs.listener, rs.Mux, rs.log, cfg); err != nil {
			rs.log.Info(fmt.Sprintf("Application REST service stopped: %v", err))
		}
	}()

	return nil
}

func (rs *StoppableRestServer) Stop() {
	rs.listener.Close()
}

// REST endpoint error object
type RestError struct {
	Error string `json:"error"`
}

// ABCI error object helper, used to unmarshal RestError.Error string
type ABCIError struct {
	Codespace sdk.CodespaceType `json:"codespace"`
	Code      sdk.CodeType      `json:"code"`
	Message   string            `json:"message"`
}

// Type that combines an Address with the privKey and pubKey to that address
type AddrKeys struct {
	Address sdk.AccAddress
	PubKey  crypto.PubKey
	PrivKey crypto.PrivKey
}

func NewAddrKeys(address sdk.AccAddress, pubKey crypto.PubKey,
	privKey crypto.PrivKey) AddrKeys {

	return AddrKeys{
		Address: address,
		PubKey:  pubKey,
		PrivKey: privKey,
	}
}

// implement `Interface` in sort package.
type AddrKeysSlice []AddrKeys

func (b AddrKeysSlice) Len() int {
	return len(b)
}

// Sorts lexographically by Address
func (b AddrKeysSlice) Less(i, j int) bool {
	// bytes package already implements Comparable for []byte.
	switch bytes.Compare(b[i].Address.Bytes(), b[j].Address.Bytes()) {
	case -1:
		return true
	case 0, 1:
		return false
	default:
		panic("not fail-able with `bytes.Comparable` bounded [-1, 1].")
	}
}

func (b AddrKeysSlice) Swap(i, j int) {
	b[j], b[i] = b[i], b[j]
}

// CreateGenAccounts generates genesis accounts loaded with coins, and returns
// their addresses, pubkeys, and privkeys.
func CreateGenAccounts(numAccs int, genCoins sdk.Coins) (genAccs []*auth.BaseAccount,
	addrs []sdk.AccAddress, pubKeys []crypto.PubKey, privKeys []crypto.PrivKey) {

	addrKeysSlice := AddrKeysSlice{}

	for i := 0; i < numAccs; i++ {
		privKey := secp256k1.GenPrivKey()
		pubKey := privKey.PubKey()
		addr := sdk.AccAddress(pubKey.Address())

		addrKeysSlice = append(addrKeysSlice, NewAddrKeys(addr, pubKey, privKey))
	}

	sort.Sort(addrKeysSlice)

	for i := range addrKeysSlice {
		addrs = append(addrs, addrKeysSlice[i].Address)
		pubKeys = append(pubKeys, addrKeysSlice[i].PubKey)
		privKeys = append(privKeys, addrKeysSlice[i].PrivKey)
		genAccs = append(genAccs, &auth.BaseAccount{
			AccountNumber: uint64(i),
			Address:       addrKeysSlice[i].Address,
			Coins:         genCoins,
			PubKey:        addrKeysSlice[i].PubKey,
		})
	}

	return
}

const (
	DefaultMockVMAddress  = "127.0.0.1:0" // Default virtual machine address to connect from Cosmos SDK.
	DefaultMockDataListen = "127.0.0.1:0" // Default data server address to listen for connections from VM.

	FlagVMMockAddress = "vm.mock.address"
	FlagDSMockListen  = "ds.mock.listen"
)

var (
	vmMockAddress  *string
	dataListenMock *string
)

func MockVMConfig() *vmConfig.VMConfig {
	return &vmConfig.VMConfig{
		Address:    *vmMockAddress,
		DataListen: *dataListenMock,
	}
}

func init() {
	if flag.Lookup(FlagVMMockAddress) == nil {
		vmMockAddress = flag.String(FlagVMMockAddress, DefaultMockVMAddress, "mocked address of virtual machine server client/server")
	}

	if flag.Lookup(FlagDSMockListen) == nil {
		dataListenMock = flag.String(FlagDSMockListen, DefaultMockDataListen, "address of mocked data server to launch/connect")
	}
}

type VMServer struct {
	vm.UnimplementedVMServiceServer
}

func newTestWbApp() (*WbServiceApp, *grpc.Server) {
	config := MockVMConfig()

	vmListener, err := net.Listen("tcp", config.Address)
	if err != nil {
		panic(err)
	}
	vmServer := VMServer{}
	server := grpc.NewServer()

	vm.RegisterVMServiceServer(server, &vmServer)

	go func() {
		if err := server.Serve(vmListener); err != nil {
			panic(err)
		}
	}()

	return NewWbServiceApp(log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "sdk/app"), dbm.NewMemDB(), config), server
}

func getGenesis(app *WbServiceApp, chainID, monikerID string, accs []*auth.BaseAccount, privValidatorKey *ed25519.PrivKeyEd25519) ([]byte, error) {
	// generate node validator account
	var genTxAcc *auth.BaseAccount
	var genTxPubKey crypto.PubKey
	var genTxPrivKey secp256k1.PrivKeySecp256k1
	{
		accCoins, _ := sdk.ParseCoins("1000000000000000wings")

		if privValidatorKey == nil {
			k := ed25519.GenPrivKey()
			privValidatorKey = &k
		}

		genTxPrivKey = secp256k1.GenPrivKey()
		genTxPubKey = genTxPrivKey.PubKey()

		accAddr := sdk.AccAddress(genTxPubKey.Address())
		genTxAcc = &auth.BaseAccount{
			AccountNumber: uint64(len(accs)),
			Address:       accAddr,
			Coins:         accCoins,
			PubKey:        privValidatorKey.PubKey(),
		}
	}

	// generate genesis state based on defaults
	genesisState := ModuleBasics.DefaultGenesis()
	{
		accounts := make(genaccounts.GenesisAccounts, 0, len(accs)+1)
		for _, acc := range accs {
			accounts = append(accounts, genaccounts.NewGenesisAccount(acc))
		}
		accounts = append(accounts, genaccounts.NewGenesisAccount(genTxAcc))

		genesisState[genaccounts.ModuleName] = codec.MustMarshalJSONIndent(app.cdc, accounts)

		validators := make(poaTypes.Validators, len(accs))
		for idx, acc := range accs {
			validators[idx] = poaTypes.Validator{Address: acc.Address, EthAddress: "0x17f7D1087971dF1a0E6b8Dae7428E97484E32615"}
		}
		genesisState[poaTypes.ModuleName] = codec.MustMarshalJSONIndent(app.cdc, poaTypes.GenesisState{
			Parameters:    poaTypes.DefaultParams(),
			PoAValidators: validators,
		})

		genesisState[msTypes.ModuleName] = codec.MustMarshalJSONIndent(app.cdc, msTypes.GenesisState{
			Parameters: msTypes.Params{
				IntervalToExecute: 50,
			},
		})

		stakingGenesis := staking.GenesisState{}
		app.cdc.MustUnmarshalJSON(genesisState[staking.ModuleName], &stakingGenesis)
		stakingGenesis.Params.BondDenom = wbConfig.MainDenom
		genesisState[staking.ModuleName] = codec.MustMarshalJSONIndent(app.cdc, stakingGenesis)
	}

	// generate node validator genTx and update genutil module genesis
	{
		commissionRate, _ := sdk.NewDecFromStr("0.100000000000000000")
		commissionMaxRate, _ := sdk.NewDecFromStr("0.200000000000000000")
		commissionChangeRate, _ := sdk.NewDecFromStr("0.010000000000000000")
		tokenAmount := sdk.TokensFromConsensusPower(500000)

		msg := staking.NewMsgCreateValidator(
			genTxAcc.Address.Bytes(),
			genTxAcc.PubKey,
			sdk.NewCoin(wbConfig.MainDenom, tokenAmount),
			staking.NewDescription(monikerID, "", "", ""),
			staking.NewCommissionRates(commissionRate, commissionMaxRate, commissionChangeRate),
			sdk.OneInt(),
		)

		txFee := auth.StdFee{
			Amount: sdk.Coins{{Denom: "wings", Amount: sdk.NewInt(1)}},
			Gas:    200000,
		}
		txMemo := "testmemo"

		signature, err := genTxPrivKey.Sign(auth.StdSignBytes(chainID, 0, 0, txFee, []sdk.Msg{msg}, txMemo))
		if err != nil {
			return nil, err
		}

		stdSig := auth.StdSignature{
			PubKey:    genTxPubKey,
			Signature: signature,
		}

		tx := auth.NewStdTx([]sdk.Msg{msg}, txFee, []auth.StdSignature{stdSig}, txMemo)

		genesisState[genutil.ModuleName] = codec.MustMarshalJSONIndent(app.cdc, genutil.NewGenesisStateFromStdTx([]auth.StdTx{tx}))
	}

	if err := ModuleBasics.ValidateGenesis(genesisState); err != nil {
		return nil, err
	}

	stateBytes, err := codec.MarshalJSONIndent(app.cdc, genesisState)
	if err != nil {
		return nil, err
	}

	return stateBytes, nil
}

func setGenesis(t *testing.T, app *WbServiceApp, accs []*auth.BaseAccount) (sdk.Context, error) {
	ctx := app.NewContext(true, abci.Header{})

	stateBytes, err := getGenesis(app, "", "testMoniker", accs, nil)
	if err != nil {
		return ctx, err
	}

	// initialize the chain
	app.InitChain(
		abci.RequestInitChain{
			AppStateBytes: stateBytes,
		},
	)
	app.Commit()

	return ctx, nil
}

func newTestWbAppWithRest(t *testing.T, genValidators []*auth.BaseAccount) (app *WbServiceApp, chainId string, stopFunc func()) {
	chainId = "wd-test"

	var rootDir string
	var node *tmNode.Node
	var restServer *StoppableRestServer
	var err error

	stopFunc = func() {
		if rootDir != "" {
			os.RemoveAll(rootDir)
		}
		if node != nil {
			node.Stop()
		}
		if restServer != nil {
			restServer.Stop()
		}
	}

	config := sdk.GetConfig()
	wbConfig.InitBechPrefixes(config)
	// config not sealed by intention: multiple test runs fail with assert on sealed

	// tmp dir primary used for "cs.wal" file (consensus write ahead logs)
	rootDir, err = ioutil.TempDir("/tmp", "wd-test-")
	require.NoError(t, err, "TempDir")

	// adjust default config
	ctx := server.NewDefaultContext()
	cfg := ctx.Config
	cfg.SetRoot(rootDir)
	cfg.Instrumentation.Prometheus = false
	cfg.Moniker = "wd-test-moniker"
	cfg.LogLevel = "main:error,state:error,*:error"

	// lower default logger filter level
	logger, err := tmFlags.ParseLogLevel(cfg.LogLevel, ctx.Logger, "error")
	require.NoError(t, err, "logger filter")
	ctx.Logger = logger

	// init the app
	db := tmDb.NewDB("appInMemDb", tmDb.MemDBBackend, "")
	app = NewWbServiceApp(ctx.Logger, db, MockVMConfig())

	privValidatorKey := ed25519.GenPrivKey()
	privValidatorFile := &privval.FilePV{
		Key: privval.FilePVKey{
			Address: privValidatorKey.PubKey().Address(),
			PubKey:  privValidatorKey.PubKey(),
			PrivKey: privValidatorKey,
		},
		LastSignState: privval.FilePVLastSignState{
			Step: 0,
		},
	}

	// generate test app state (genesis)
	appState, err := getGenesis(app, chainId, cfg.Moniker, genValidators, &privValidatorKey)
	require.NoError(t, err, "getGenesis")

	// tendermint node configuration
	privNodeKey := ed25519.GenPrivKey()
	nodeKey := &p2p.NodeKey{PrivKey: privNodeKey}

	genesisDocProvider := func() (*tmTypes.GenesisDoc, error) {
		genDoc := &tmTypes.GenesisDoc{ChainID: chainId, AppState: appState}
		return genDoc, nil
	}

	dbProvider := func(*tmNode.DBContext) (tmDbm.DB, error) {
		return tmDb.NewDB("nodeInMemDb", tmDb.MemDBBackend, ""), nil
	}

	// tendermint node start
	node, err = tmNode.NewNode(
		cfg,
		privValidatorFile,
		nodeKey,
		proxy.NewLocalClientCreator(app),
		genesisDocProvider,
		dbProvider,
		tmNode.DefaultMetricsProvider(cfg.Instrumentation),
		ctx.Logger.With("module", "node"),
	)
	require.NoError(t, err, "node.NewNode")
	require.NoError(t, node.Start(), "node.Start")

	// commit genesis
	app.Commit()

	// start REST server
	restServer = NewStoppableRestServer(app.cdc, rpcClient.NewLocal(node))
	require.NoError(t, restServer.Start("tcp://localhost:1317", 10, 5, 5), "server start")

	return
}

func RestRequest(t *testing.T, app *WbServiceApp, httpMethod, urlSubPath string, urlValues url.Values, requestValue interface{}, responseValue interface{}, doCheck bool) (retCode int, retErrBody []byte) {
	u, _ := url.Parse("http://localhost:1317")
	u.Path = path.Join(u.Path, urlSubPath)
	if urlValues != nil {
		u.RawQuery = urlValues.Encode()
	}

	_, err := url.Parse(u.String())
	require.NoError(t, err, "ParseRequestURI: %s", u.String())

	var reqBodyBytes []byte
	if requestValue != nil {
		var err error
		reqBodyBytes, err = app.cdc.MarshalJSON(requestValue)
		require.NoError(t, err, "requestValue")
	}

	req, err := http.NewRequest(httpMethod, u.String(), bytes.NewBuffer(reqBodyBytes))
	require.NoError(t, err, "NewRequest")
	req.Header.Set("Content-type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	require.NoError(t, err, "HTTP request")
	require.NotNil(t, resp, "HTTP response")

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err, "HTTP response body read")

	retCode = resp.StatusCode
	if doCheck {
		require.Equal(t, retCode, http.StatusOK, "HTTP code %q (%s): %s", resp.Status, u.String(), string(bodyBytes))
	}

	if retCode != http.StatusOK {
		retErrBody = bodyBytes
		app.cdc.UnmarshalJSON(bodyBytes, responseValue)

		return
	}

	if responseValue != nil {
		respMsg := sdkRest.ResponseWithHeight{}
		require.NoError(t, app.cdc.UnmarshalJSON(bodyBytes, &respMsg), "ResponseWithHeight")
		require.NoError(t, app.cdc.UnmarshalJSON(respMsg.Result, responseValue), "responseValue")
	}

	return
}

func CheckRestError(t *testing.T, app *WbServiceApp, expectedCode, receivedCode int, expectedErr sdk.Error, receivedBody []byte) {
	require.Equal(t, expectedCode, receivedCode, "code")

	if expectedErr != nil {
		require.NotNil(t, receivedBody, "receivedBody")

		restErr, abciErr := &RestError{}, &ABCIError{}
		require.NoError(t, app.cdc.UnmarshalJSON(receivedBody, restErr), "unmarshal to RestError: %s", string(receivedBody))
		require.NoError(t, app.cdc.UnmarshalJSON([]byte(restErr.Error), abciErr), "unmarshal to ABCIError: %s", string(receivedBody))
		require.Equal(t, expectedErr.Codespace(), abciErr.Codespace, "Codespace: %s", string(receivedBody))
		require.Equal(t, expectedErr.Code(), abciErr.Code, "Code: %s", string(receivedBody))
	}
}

// GenTx generates a signed mock transaction.
func genTx(msgs []sdk.Msg, accnums []uint64, seq []uint64, priv ...crypto.PrivKey) auth.StdTx {
	sigs := make([]auth.StdSignature, len(priv))
	memo := "testmemotestmemo"

	fee := auth.StdFee{
		Amount: sdk.Coins{{Denom: "wings", Amount: sdk.NewInt(1)}},
		Gas:    200000,
	}

	for i, p := range priv {
		sig, err := p.Sign(auth.StdSignBytes(chainID, accnums[i], seq[i], fee, msgs, memo))
		if err != nil {
			panic(err)
		}

		sigs[i] = auth.StdSignature{
			PubKey:    p.PubKey(),
			Signature: sig,
		}
	}

	return auth.NewStdTx(msgs, fee, sigs, memo)
}

func DeliverTx(app *WbServiceApp, tx auth.StdTx) sdk.Result {
	if res := app.Simulate(app.cdc.MustMarshalJSON(tx), tx); !res.IsOK() {
		return res
	}

	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{ChainID: chainID, Height: app.LastBlockHeight() + 1}})
	if res := app.Deliver(tx); !res.IsOK() {
		return res
	}
	app.EndBlock(abci.RequestEndBlock{})
	app.Commit()

	return sdk.Result{}
}

// DeliverTx and success check
func CheckDeliverTx(t *testing.T, app *WbServiceApp, tx auth.StdTx) {
	res := DeliverTx(app, tx)
	require.True(t, res.IsOK(), res.Log)
}

// DeliverTx and fail check
func CheckDeliverErrorTx(t *testing.T, app *WbServiceApp, tx auth.StdTx) {
	res := DeliverTx(app, tx)
	require.False(t, res.IsOK(), res.Log)
}

// DeliverTx and fail check with specific error
func CheckDeliverSpecificErrorTx(t *testing.T, app *WbServiceApp, tx auth.StdTx, err sdk.Error) {
	res := DeliverTx(app, tx)
	CheckResultError(t, err, res)
}

func RunQuery(t *testing.T, app *WbServiceApp, requestData interface{}, path string, responseValue interface{}) abci.ResponseQuery {
	resp := app.Query(abci.RequestQuery{
		Data: codec.MustMarshalJSONIndent(app.cdc, requestData),
		Path: path,
	})

	if responseValue != nil && resp.IsOK() {
		require.NoError(t, app.cdc.UnmarshalJSON(resp.Value, responseValue))
	}

	return resp
}

// RunQuery and success check
func CheckRunQuery(t *testing.T, app *WbServiceApp, requestData interface{}, path string, responseValue interface{}) {
	resp := RunQuery(t, app, requestData, path, responseValue)
	require.True(t, resp.IsOK())
}

// RunQuery and fail check with specific error
func CheckRunQuerySpecificError(t *testing.T, app *WbServiceApp, requestData interface{}, path string, err sdk.Error) {
	resp := RunQuery(t, app, requestData, path, nil)
	require.True(t, resp.IsErr())
	require.Equal(t, string(err.Codespace()), resp.Codespace, "Codespace: %s", resp.Log)
	require.Equal(t, uint32(err.Code()), resp.Code, "Code: %s", resp.Log)
}

func GetContext(app *WbServiceApp, isCheckTx bool) sdk.Context {
	return app.NewContext(isCheckTx, abci.Header{Height: app.LastBlockHeight() + 1})
}

func GetAccountCheckTx(app *WbServiceApp, address sdk.AccAddress) auth.Account {
	return app.accountKeeper.GetAccount(app.NewContext(true, abci.Header{Height: app.LastBlockHeight() + 1}), address)
}

func GetAccount(app *WbServiceApp, address sdk.AccAddress) auth.Account {
	return app.accountKeeper.GetAccount(app.NewContext(false, abci.Header{Height: app.LastBlockHeight() + 1}), address)
}

// Check if expected / received tx results are equal
func CheckResultError(t *testing.T, expectedErr sdk.Error, receivedRes sdk.Result) {
	require.Equal(t, expectedErr.Codespace(), receivedRes.Codespace, "Codespace: %s", receivedRes.Log)
	require.Equal(t, expectedErr.Code(), receivedRes.Code, "Code: %s", receivedRes.Log)
}

func MSMsgSubmitAndVote(t *testing.T, app *WbServiceApp, msMsgID string, msMsg core.MsMsg, submitAccIdx uint, accs []*auth.BaseAccount, privKeys []crypto.PrivKey, doChecks bool) sdk.Result {
	confirmCnt := int(app.poaKeeper.GetEnoughConfirmations(GetContext(app, true)))

	// lazy input check
	require.Equal(t, len(accs), len(privKeys), "invalid input: accs / privKeys len mismatch")
	require.Less(t, submitAccIdx, uint(len(accs)), "invalid input: submitAccIdx >= len(accs)")
	require.Less(t, submitAccIdx, uint(len(accs)), "invalid input: submitAccIdx >= len(accs)")
	require.LessOrEqual(t, confirmCnt, len(accs), "invalid input: confirmations count > len(accs)")

	callMsgID := uint64(0)
	{
		// submit message
		senderAcc, senderPrivKey := GetAccountCheckTx(app, accs[submitAccIdx].Address), privKeys[submitAccIdx]
		submitMsg := msMsgs.NewMsgSubmitCall(msMsg, msMsgID, senderAcc.GetAddress())
		tx := genTx([]sdk.Msg{submitMsg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
		if doChecks {
			CheckDeliverTx(t, app, tx)
		} else if res := DeliverTx(app, tx); !res.IsOK() {
			return res
		}

		// check vote added
		calls := msTypes.CallsResp{}
		CheckRunQuery(t, app, nil, queryMsGetCallsPath, &calls)
		require.Equal(t, 1, len(calls[0].Votes))

		callMsgID = calls[0].Call.MsgID
	}

	// cut submit message sender from accounts
	accsFixed, privKeysFixed := append([]*auth.BaseAccount(nil), accs...), append([]crypto.PrivKey(nil), privKeys...)
	accsFixed = append(accsFixed[:submitAccIdx], accsFixed[submitAccIdx+1:]...)
	privKeysFixed = append(privKeysFixed[:submitAccIdx], privKeysFixed[submitAccIdx+1:]...)

	// voting (confirming)
	for idx := 0; idx < confirmCnt-2; idx++ {
		// confirm message
		senderAcc, senderPrivKey := GetAccountCheckTx(app, accsFixed[idx].Address), privKeysFixed[idx]
		confirmMsg := msMsgs.NewMsgConfirmCall(callMsgID, senderAcc.GetAddress())
		tx := genTx([]sdk.Msg{confirmMsg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
		if doChecks {
			CheckDeliverTx(t, app, tx)
		} else if res := DeliverTx(app, tx); !res.IsOK() {
			return res
		}

		// check vote added / call removed
		calls := msTypes.CallsResp{}
		CheckRunQuery(t, app, nil, queryMsGetCallsPath, &calls)
		require.Equal(t, idx+2, len(calls[0].Votes))
	}

	// voting (last confirm)
	{
		// confirm message
		idx := len(accsFixed) - 1
		senderAcc, senderPrivKey := GetAccountCheckTx(app, accsFixed[idx].Address), privKeysFixed[idx]
		confirmMsg := msMsgs.NewMsgConfirmCall(callMsgID, senderAcc.GetAddress())
		tx := genTx([]sdk.Msg{confirmMsg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
		if doChecks {
			CheckDeliverTx(t, app, tx)
		} else if res := DeliverTx(app, tx); !res.IsOK() {
			return res
		}

		// check call removed
		calls := msTypes.CallsResp{}
		CheckRunQuery(t, app, nil, queryMsGetCallsPath, &calls)
		require.Equal(t, 0, len(calls))
	}

	return sdk.Result{}
}
