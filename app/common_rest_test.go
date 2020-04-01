package app

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdkKeybase "github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkRest "github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/auth"
	sdkAuthRest "github.com/cosmos/cosmos-sdk/x/auth/client/rest"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	tmFlags "github.com/tendermint/tendermint/libs/cli/flags"
	"github.com/tendermint/tendermint/libs/log"
	tmNode "github.com/tendermint/tendermint/node"
	"github.com/tendermint/tendermint/p2p"
	"github.com/tendermint/tendermint/privval"
	"github.com/tendermint/tendermint/proxy"
	rpcClient "github.com/tendermint/tendermint/rpc/client"
	tmCoreTypes "github.com/tendermint/tendermint/rpc/core/types"
	rpcServer "github.com/tendermint/tendermint/rpc/lib/server"
	tmTypes "github.com/tendermint/tendermint/types"
	tmDb "github.com/tendermint/tm-db"
	tmDbm "github.com/tendermint/tm-db"

	dnConfig "github.com/dfinance/dnode/cmd/config"
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

func NewStoppableRestServer(cdc *codec.Codec, customRpcClient rpcClient.Client, printLogs bool) *StoppableRestServer {
	r := mux.NewRouter()
	cliCtx := context.NewCLIContext().WithCodec(cdc).WithTrustNode(true).WithClient(customRpcClient)

	var logger log.Logger
	if printLogs {
		logger = log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "rest-server")
	} else {
		logger = log.NewNopLogger()
	}

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

	cfg := rpcServer.DefaultConfig()
	cfg.MaxOpenConnections = maxOpen
	cfg.ReadTimeout = time.Duration(readTimeout) * time.Second
	cfg.WriteTimeout = time.Duration(writeTimeout) * time.Second

	rs.listener, err = rpcServer.Listen(listenAddr, cfg)
	if err != nil {
		return
	}
	rs.log.Info("Starting application REST service...")

	go func() {
		if err := rpcServer.StartHTTPServer(rs.listener, rs.Mux, rs.log, cfg); err != nil {
			rs.log.Info(fmt.Sprintf("Application REST service stopped: %v", err))
		}
	}()

	return nil
}

func (rs *StoppableRestServer) Stop() {
	rs.listener.Close()
}

type RestTester struct {
	RootDir          string
	ChainId          string
	DefaultAssetCode string
	Accounts         []*auth.BaseAccount
	PrivKeys         []crypto.PrivKey
	t                *testing.T
	App              *DnServiceApp
	node             *tmNode.Node
	restServer       *StoppableRestServer
}

func NewRestTester(t *testing.T, printLogs bool) *RestTester {
	var err error

	r := RestTester{
		ChainId:          "dn-test",
		DefaultAssetCode: "oracle_rest_asset",
		t:                t,
	}

	genAccs, _, _, genPrivKeys := CreateGenAccounts(7, GenDefCoins(t))
	r.Accounts, r.PrivKeys = genAccs, genPrivKeys

	// sdk.GetConfig() setup and seal is omitted as oracle-app does it at the init() stage already

	// tmp dir primary used for "cs.wal" file (consensus write ahead logs)
	r.RootDir, err = ioutil.TempDir("/tmp", "wd-test-")
	require.NoError(r.t, err, "TempDir")

	require.NoError(r.t, os.MkdirAll(path.Join(r.RootDir, "config"), 0755), "ConfigDir")
	require.NoError(r.t, os.MkdirAll(path.Join(r.RootDir, "data"), 0755), "DataDir")

	// adjust default config
	ctx := server.NewDefaultContext()
	cfg := ctx.Config
	cfg.SetRoot(r.RootDir)
	cfg.Instrumentation.Prometheus = false
	cfg.Moniker = "dn-test-moniker"
	cfg.LogLevel = "*:none"
	if printLogs {
		cfg.LogLevel = "main:error,state:error,*:error"
	}

	// lower default logger filter level
	logger, err := tmFlags.ParseLogLevel(cfg.LogLevel, ctx.Logger, "error")
	require.NoError(r.t, err, "logger filter")
	ctx.Logger = logger

	// init the app
	db := tmDb.NewDB("appInMemDb", tmDb.MemDBBackend, "")
	r.App = NewDnServiceApp(ctx.Logger, db, MockVMConfig())

	//privValidatorKey := ed25519.GenPrivKey()
	//privValidatorFile := &privval.FilePV{
	//	Key: privval.FilePVKey{
	//		Address: privValidatorKey.PubKey().Address(),
	//		PubKey:  privValidatorKey.PubKey(),
	//		PrivKey: privValidatorKey,
	//	},
	//	LastSignState: privval.FilePVLastSignState{
	//		Step: 0,
	//	},
	//}
	privValidatorFile := privval.GenFilePV(cfg.PrivValidatorKeyFile(), cfg.PrivValidatorStateFile())
	privValidatorFile.Save()
	pvKeyED25519 := privValidatorFile.Key.PrivKey.(ed25519.PrivKeyEd25519)

	// generate test app state (genesis)
	appState, err := getGenesis(r.App, r.ChainId, cfg.Moniker, r.Accounts, &pvKeyED25519)
	require.NoError(r.t, err, "getGenesis")

	// tendermint node configuration
	//nodeKey, err := p2p.LoadOrGenNodeKey(cfg.NodeKeyFile())
	//require.NoError(r.t, err, "LoadOrGenNodeKey")
	privNodeKey := ed25519.GenPrivKey()
	nodeKey := &p2p.NodeKey{PrivKey: privNodeKey}

	genesisDocProvider := func() (*tmTypes.GenesisDoc, error) {
		genDoc := &tmTypes.GenesisDoc{ChainID: r.ChainId, AppState: appState}
		return genDoc, nil
	}

	dbProvider := func(*tmNode.DBContext) (tmDbm.DB, error) {
		return tmDb.NewDB("nodeInMemDb", tmDb.MemDBBackend, ""), nil
	}

	// tendermint node start
	r.node, err = tmNode.NewNode(
		cfg,
		privValidatorFile,
		nodeKey,
		proxy.NewLocalClientCreator(r.App),
		genesisDocProvider,
		dbProvider,
		tmNode.DefaultMetricsProvider(cfg.Instrumentation),
		ctx.Logger.With("module", "node"),
	)
	require.NoError(r.t, err, "node.NewNode")
	require.NoError(r.t, r.node.Start(), "node.Start")

	// start REST server
	r.restServer = NewStoppableRestServer(r.App.cdc, rpcClient.NewLocal(r.node), printLogs)
	require.NoError(r.t, r.restServer.Start("tcp://localhost:1317", 10, 5, 5), "server start")

	// wait for node to start
	r.WaitForBlockHeight(2)

	return &r
}

func (r *RestTester) Close() {
	if r.restServer != nil {
		r.restServer.Stop()
	}
	if r.node != nil {
		r.node.Stop()
	}
	if r.RootDir != "" {
		os.RemoveAll(r.RootDir)
	}
	time.Sleep(1 * time.Second)
}

// Get current block height
func (r *RestTester) CurrentBlockHeight() int64 {
	res, err := http.Get("http://localhost:1317/blocks/latest")
	require.NoError(r.t, err, "blocks/latest")

	body, err := ioutil.ReadAll(res.Body)
	require.NoError(r.t, err, "reading body")
	require.NoError(r.t, res.Body.Close(), "closing body")

	resultBlock := tmCoreTypes.ResultBlock{}
	require.NoError(r.t, r.App.cdc.UnmarshalJSON(body, &resultBlock), "unmarshal body")

	if resultBlock.Block == nil {
		return 0
	}

	return resultBlock.Block.Height
}

// Wait until node reaches {blockHeight}
func (r *RestTester) WaitForBlockHeight(targetHeight int64) {
	for {
		curHeight := r.CurrentBlockHeight()
		if curHeight >= targetHeight {
			return
		}

		time.Sleep(time.Millisecond * 100)
	}
}

// Wait until node starts a new block
func (r *RestTester) WaitForNextBlock() int64 {
	prevHeight := r.CurrentBlockHeight()
	for {
		time.Sleep(time.Millisecond * 50)

		curHeight := r.CurrentBlockHeight()
		if curHeight != prevHeight {
			return curHeight
		}
	}
}

// Send REST request with relative subPath, URL variables and optional request/response (pointer) objects
//   {doCheck} flag checks if request was successful
func (r *RestTester) Request(httpMethod, urlSubPath string, urlValues url.Values, requestValue interface{}, responseValue interface{}, doCheck bool) (retCode int, retErrBody []byte) {
	u, _ := url.Parse("http://localhost:1317")
	u.Path = path.Join(u.Path, urlSubPath)
	if urlValues != nil {
		u.RawQuery = urlValues.Encode()
	}

	_, err := url.Parse(u.String())
	require.NoError(r.t, err, "ParseRequestURI: %s", u.String())

	var reqBodyBytes []byte
	if requestValue != nil {
		var err error
		reqBodyBytes, err = r.App.cdc.MarshalJSON(requestValue)
		require.NoError(r.t, err, "requestValue")
	}

	req, err := http.NewRequest(httpMethod, u.String(), bytes.NewBuffer(reqBodyBytes))
	require.NoError(r.t, err, "NewRequest")
	req.Header.Set("Content-type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(r.t, err, "HTTP request")
	require.NotNil(r.t, resp, "HTTP response")
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	require.NoError(r.t, err, "HTTP response body read")

	retCode = resp.StatusCode
	if doCheck {
		require.Equal(r.t, retCode, http.StatusOK, "HTTP code %q (%s): %s", resp.Status, u.String(), string(bodyBytes))
	}

	if retCode != http.StatusOK {
		retErrBody = bodyBytes
		r.App.cdc.UnmarshalJSON(bodyBytes, responseValue)

		return
	}

	// parse Tx response or Query response
	if responseValue != nil {
		if _, ok := responseValue.(*sdk.TxResponse); !ok {
			respMsg := sdkRest.ResponseWithHeight{}
			require.NoError(r.t, r.App.cdc.UnmarshalJSON(bodyBytes, &respMsg), "ResponseWithHeight: %s", string(bodyBytes))
			if respMsg.Result != nil {
				require.NoError(r.t, r.App.cdc.UnmarshalJSON(respMsg.Result, responseValue), "responseValue: %s", string(bodyBytes))
			}
		} else {
			require.NoError(r.t, r.App.cdc.UnmarshalJSON(bodyBytes, responseValue), "txResponseValue: %s", string(bodyBytes))
		}
	}

	return
}

// Prepare and REST send Tx from RestTester.Accounts[{senderIdx}] embedding {msg}
//   {doCheck} checks if Tx was submitted
//   {sync} true - syncTx / false - blockTx
func (r *RestTester) txRequest(senderIdx uint, msg sdk.Msg, sync, doCheck bool) sdk.TxResponse {
	senderAddr, senderPrivKey := r.Accounts[senderIdx].Address, r.PrivKeys[senderIdx]

	// build broadcast Tx request
	senderAcc := GetAccountCheckTx(r.App, senderAddr)

	txFee := auth.StdFee{
		Amount: sdk.Coins{{Denom: dnConfig.MainDenom, Amount: sdk.NewInt(1)}},
		Gas:    200000,
	}
	txMemo := "restTxMemo"

	signature, err := senderPrivKey.Sign(auth.StdSignBytes(r.ChainId, senderAcc.GetAccountNumber(), senderAcc.GetSequence(), txFee, []sdk.Msg{msg}, txMemo))
	require.NoError(r.t, err, "signing Tx")

	stdSig := auth.StdSignature{
		PubKey:    senderPrivKey.PubKey(),
		Signature: signature,
	}
	tx := auth.NewStdTx([]sdk.Msg{msg}, txFee, []auth.StdSignature{stdSig}, txMemo)

	txBroadcastReq := sdkAuthRest.BroadcastReq{
		Tx:   tx,
		Mode: "block",
	}

	if sync {
		txBroadcastReq.Mode = "sync"
	}

	// send Tx
	txResp := sdk.TxResponse{}
	r.Request("POST", "txs", nil, txBroadcastReq, &txResp, true)

	// check if Tx successful
	if !doCheck {
		return txResp
	}
	require.Equal(r.t, sdk.CodeOK, sdk.CodeType(txResp.Code), "tx failed: %v", txResp)

	return txResp
}

func (r *RestTester) TxBlockRequest(senderIdx uint, msg sdk.Msg, doCheck bool) sdk.TxResponse {
	return r.txRequest(senderIdx, msg, false, doCheck)
}

func (r *RestTester) TxSyncRequest(senderIdx uint, msg sdk.Msg, doCheck bool) sdk.TxResponse {
	return r.txRequest(senderIdx, msg, true, doCheck)
}

func (r *RestTester) CheckError(expectedCode, receivedCode int, expectedErr sdk.Error, receivedBody []byte) {
	require.Equal(r.t, expectedCode, receivedCode, "code")

	if expectedErr != nil {
		require.NotNil(r.t, receivedBody, "receivedBody")

		restErr, abciErr := &RestError{}, &ABCIError{}
		require.NoError(r.t, r.App.cdc.UnmarshalJSON(receivedBody, restErr), "unmarshal to RestError: %s", string(receivedBody))
		require.NoError(r.t, r.App.cdc.UnmarshalJSON([]byte(restErr.Error), abciErr), "unmarshal to ABCIError: %s", string(receivedBody))
		require.Equal(r.t, expectedErr.Codespace(), abciErr.Codespace, "Codespace: %s", string(receivedBody))
		require.Equal(r.t, expectedErr.Code(), abciErr.Code, "Code: %s", string(receivedBody))
	}
}
