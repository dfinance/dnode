// +build unit

package app

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"sort"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authExported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"

	dnConfig "github.com/dfinance/dnode/cmd/config"
	vmConfig "github.com/dfinance/dnode/cmd/config"
	"github.com/dfinance/dnode/x/core"
	"github.com/dfinance/dnode/x/genaccounts"
	msMsgs "github.com/dfinance/dnode/x/multisig/msgs"
	msTypes "github.com/dfinance/dnode/x/multisig/types"
	"github.com/dfinance/dnode/x/oracle"
	poaTypes "github.com/dfinance/dnode/x/poa/types"
	"github.com/dfinance/dnode/x/vm"
	"github.com/dfinance/dnode/x/vmauth"
)

var (
	chainID         = ""
	currency1Symbol = "testa"
	currency2Symbol = "testb"
	currency3Symbol = "testc"
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
	bufferSize = 1024 * 1024
)

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

	if err := vmauth.AddDenomPath(currency1Symbol, "00"); err != nil {
		panic(err)
	}

	if err := vmauth.AddDenomPath(currency2Symbol, "01"); err != nil {
		panic(err)
	}

	if err := vmauth.AddDenomPath(currency3Symbol, "11"); err != nil {
		panic(err)
	}
}

type VMServer struct {
	vm.UnimplementedVMServiceServer
}

func newTestDnApp() (*DnServiceApp, *grpc.Server) {
	config := MockVMConfig()

	vmListener := bufconn.Listen(bufferSize)

	vmServer := VMServer{}
	server := grpc.NewServer()

	vm.RegisterVMServiceServer(server, &vmServer)

	go func() {
		if err := server.Serve(vmListener); err != nil {
			panic(err)
		}
	}()

	return NewDnServiceApp(log.NewTMLogger(log.NewSyncWriter(os.Stdout)).With("module", "sdk/app"), dbm.NewMemDB(), config), server
}

func getGenesis(app *DnServiceApp, chainID, monikerID string, accs []*auth.BaseAccount, privValidatorKey *ed25519.PrivKeyEd25519) ([]byte, error) {
	// generate node validator account
	var genTxAcc *auth.BaseAccount
	var genTxPubKey crypto.PubKey
	var genTxPrivKey secp256k1.PrivKeySecp256k1
	{
		genTxPrivKey = secp256k1.GenPrivKey()
		genTxPubKey = genTxPrivKey.PubKey()

		accAddr := sdk.AccAddress(genTxPubKey.Address())
		genTxAcc = &auth.BaseAccount{
			AccountNumber: uint64(len(accs)),
			Address:       accAddr,
			Coins:         GenDefCoins(nil),
			PubKey:        genTxPubKey,
		}
	}

	// generate genesis state based on defaults
	genesisState := ModuleBasics.DefaultGenesis()
	{
		accounts := make(genaccounts.GenesisState, 0, len(accs)+1)
		for _, acc := range accs {
			accounts = append(accounts, *acc)
		}
		accounts = append(accounts, *genTxAcc)

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
		stakingGenesis.Params.BondDenom = dnConfig.MainDenom
		genesisState[staking.ModuleName] = codec.MustMarshalJSONIndent(app.cdc, stakingGenesis)

		oracleGenesis := oracle.GenesisState{
			Params: oracle.Params{
				Assets: []oracle.Asset{
					{
						AssetCode: "oraclerest_asset",
						Oracles:   []oracle.Oracle{},
						Active:    true,
					},
				},
				Nominees: []string{},
			},
		}
		for i := 0; i < len(accs) && i < 2; i++ {
			oracleGenesis.Params.Assets[0].Oracles = append(oracleGenesis.Params.Assets[0].Oracles, oracle.Oracle{Address: accs[i].Address})
			oracleGenesis.Params.Nominees = append(oracleGenesis.Params.Nominees, accs[i].Address.String())
		}
		genesisState[oracle.ModuleName] = codec.MustMarshalJSONIndent(app.cdc, oracleGenesis)
	}

	// generate node validator genTx and update genutil module genesis
	{
		commissionRate, _ := sdk.NewDecFromStr("0.100000000000000000")
		commissionMaxRate, _ := sdk.NewDecFromStr("0.200000000000000000")
		commissionChangeRate, _ := sdk.NewDecFromStr("0.010000000000000000")
		tokenAmount := sdk.TokensFromConsensusPower(1)

		msg := staking.NewMsgCreateValidator(
			genTxAcc.Address.Bytes(),
			genTxAcc.PubKey,
			sdk.NewCoin(dnConfig.MainDenom, tokenAmount),
			staking.NewDescription(monikerID, "", "", "", ""),
			staking.NewCommissionRates(commissionRate, commissionMaxRate, commissionChangeRate),
			sdk.OneInt(),
		)

		txFee := auth.StdFee{
			Amount: sdk.Coins{{Denom: dnConfig.MainDenom, Amount: sdk.NewInt(1)}},
			Gas:    300000,
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

func setGenesis(t *testing.T, app *DnServiceApp, accs []*auth.BaseAccount) (sdk.Context, error) {
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

// GenTx generates a signed mock transaction.
func genTx(msgs []sdk.Msg, accnums []uint64, seq []uint64, priv ...crypto.PrivKey) auth.StdTx {
	sigs := make([]auth.StdSignature, len(priv))
	memo := "testmemotestmemo"

	fee := auth.StdFee{
		Amount: sdk.Coins{{Denom: dnConfig.MainDenom, Amount: sdk.NewInt(1)}},
		Gas:    250000,
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

func GenDefCoins(t *testing.T) sdk.Coins {
	coins, err := sdk.ParseCoins("1000000000000000000000" + dnConfig.MainDenom)
	if t != nil {
		require.NoError(t, err)
	}

	return coins
}

func DeliverTx(app *DnServiceApp, tx auth.StdTx) (*sdk.Result, error) {
	if _, res, err := app.Simulate(app.cdc.MustMarshalJSON(tx), tx); err != nil {
		return res, err
	}

	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{ChainID: chainID, Height: app.LastBlockHeight() + 1}})
	_, res, err := app.Deliver(tx)
	if err != nil {
		return res, err
	}
	app.EndBlock(abci.RequestEndBlock{})
	app.Commit()

	return res, nil
}

// DeliverTx and success check
func CheckDeliverTx(t *testing.T, app *DnServiceApp, tx auth.StdTx) {
	res, err := DeliverTx(app, tx)

	resLog := ""
	if res != nil {
		resLog = res.Log
	}

	require.NoError(t, err, "res.Log %q: %v", resLog, err)
}

// DeliverTx and fail check
func CheckDeliverErrorTx(t *testing.T, app *DnServiceApp, tx auth.StdTx) {
	res, err := DeliverTx(app, tx)

	resLog := ""
	if res != nil {
		resLog = res.Log
	}

	require.Error(t, err, resLog)
}

// DeliverTx and fail check with specific error
func CheckDeliverSpecificErrorTx(t *testing.T, app *DnServiceApp, tx auth.StdTx, expectedErr error) {
	res, err := DeliverTx(app, tx)
	CheckResultError(t, expectedErr, res, err)
}

func RunQuery(t *testing.T, app *DnServiceApp, requestData interface{}, path string, responseValue interface{}) abci.ResponseQuery {
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
func CheckRunQuery(t *testing.T, app *DnServiceApp, requestData interface{}, path string, responseValue interface{}) {
	resp := RunQuery(t, app, requestData, path, responseValue)
	require.True(t, resp.IsOK())
}

// RunQuery and fail check with specific error
func CheckRunQuerySpecificError(t *testing.T, app *DnServiceApp, requestData interface{}, path string, expectedErr error) {
	expectedSdkErr, ok := expectedErr.(*sdkErrors.Error)
	require.True(t, ok, "expectedErr not a SDK error")

	resp := RunQuery(t, app, requestData, path, nil)
	require.True(t, resp.IsErr())
	require.Equal(t, expectedSdkErr.Codespace(), resp.Codespace, "Codespace: %s", resp.Log)
	require.Equal(t, expectedSdkErr.ABCICode(), resp.Code, "Code: %s", resp.Log)
}

func GetContext(app *DnServiceApp, isCheckTx bool) sdk.Context {
	return app.NewContext(isCheckTx, abci.Header{Height: app.LastBlockHeight() + 1})
}

func GetAccountCheckTx(app *DnServiceApp, address sdk.AccAddress) authExported.Account {
	return app.accountKeeper.GetAccount(app.NewContext(true, abci.Header{Height: app.LastBlockHeight() + 1}), address)
}

func GetAccount(app *DnServiceApp, address sdk.AccAddress) authExported.Account {
	return app.accountKeeper.GetAccount(app.NewContext(false, abci.Header{Height: app.LastBlockHeight() + 1}), address)
}

// Check if expected / received tx results are equal
func CheckResultError(t *testing.T, expectedErr error, receivedRes *sdk.Result, receivedErr error) {
	expectedSdkErr, ok := expectedErr.(*sdkErrors.Error)
	require.True(t, ok, "expectedErr not a SDK error: %T", expectedErr)

	resMsg := ResultErrorMsg(receivedRes, receivedErr)
	require.Error(t, receivedErr, resMsg)
	require.True(t, expectedSdkErr.Is(receivedErr), resMsg)
}

func ResultErrorMsg(res *sdk.Result, err error) string {
	resLog := ""
	if res != nil {
		resLog = res.Log
	}

	return fmt.Sprintf("result with log %q: %v", resLog, err)
}

func MSMsgSubmitAndVote(t *testing.T, app *DnServiceApp, msMsgID string, msMsg core.MsMsg, submitAccIdx uint, accs []*auth.BaseAccount, privKeys []crypto.PrivKey, doChecks bool) (*sdk.Result, error) {
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
		} else if res, err := DeliverTx(app, tx); err != nil {
			return res, err
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
		} else if res, err := DeliverTx(app, tx); err != nil {
			return res, err
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
		} else if res, err := DeliverTx(app, tx); err != nil {
			return res, err
		}

		// check call removed
		calls := msTypes.CallsResp{}
		CheckRunQuery(t, app, nil, queryMsGetCallsPath, &calls)
		require.Equal(t, 0, len(calls))
	}

	return nil, nil
}
