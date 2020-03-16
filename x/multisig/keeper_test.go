package multisig

import (
	"github.com/WingsDao/wings-blockchain/x/core"
	mstypes "github.com/WingsDao/wings-blockchain/x/multisig/types"
	"github.com/WingsDao/wings-blockchain/x/poa"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/supply"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"
	"testing"
)

const (
	ethAddress1 = "0x29D7d1dd5B6f9C864d9db560D72a247c178aE86B"
	ethAddress2 = "0x29D7d1dd5B6f9C864d9db560D72a247c178aE87B"
	ethAddress3 = "0x29D7d1dd5B6f9C864d9db560D72a247c178aE88B"
	ethAddress4 = "0x29D7d1dd5B6f9C864d9db560D72a247c178aE89B"
	//
	msgRouteNoop = "noop"
)

var (
	maccPerms = map[string][]string{
		auth.FeeCollectorName: nil,
	}
)

type testInput struct {
	cdc *codec.Codec
	ctx sdk.Context

	keyMain    *sdk.KVStoreKey
	keyAccount *sdk.KVStoreKey
	keySupply  *sdk.KVStoreKey
	keyPoa     *sdk.KVStoreKey
	keyParams  *sdk.KVStoreKey
	tkeyParams *sdk.TransientStoreKey
	keyMs      *sdk.KVStoreKey

	msRouter core.Router

	accountKeeper auth.AccountKeeper
	bankKeeper    bank.Keeper
	supplyKeeper  supply.Keeper
	paramsKeeper  params.Keeper
	poaKeeper     poa.Keeper

	target Keeper
}

func (ti *testInput) ModuleAccountAddrs() map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range maccPerms {
		modAccAddrs[supply.NewModuleAddress(acc).String()] = true
	}

	return modAccAddrs
}

type TestMsg struct {
	msgRoute string
	msgType  string
}

func (m TestMsg) Route() string            { return m.msgRoute }
func (m TestMsg) Type() string             { return m.msgType }
func (m TestMsg) ValidateBasic() sdk.Error { return nil }

func NewTestMsg(msgRoute, msgType string) TestMsg {
	return TestMsg{
		msgRoute: msgRoute,
		msgType:  msgType,
	}
}

func setupTestInput(t *testing.T) testInput {
	input := testInput{
		cdc:        codec.New(),
		keyMain:    sdk.NewKVStoreKey("main"),
		keyAccount: sdk.NewKVStoreKey(auth.StoreKey),
		keySupply:  sdk.NewKVStoreKey(supply.StoreKey),
		keyPoa:     sdk.NewKVStoreKey(poa.StoreKey),
		keyParams:  sdk.NewKVStoreKey(params.StoreKey),
		tkeyParams: sdk.NewTransientStoreKey(params.TStoreKey),
		keyMs:      sdk.NewKVStoreKey(StoreKey),
	}

	RegisterCodec(input.cdc)
	sdk.RegisterCodec(input.cdc)
	codec.RegisterCrypto(input.cdc)

	db := dbm.NewMemDB()
	mstore := store.NewCommitMultiStore(db)
	mstore.MountStoreWithDB(input.keyMain, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.keyAccount, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.keySupply, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.keyPoa, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.keyParams, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.tkeyParams, sdk.StoreTypeTransient, db)
	mstore.MountStoreWithDB(input.keyMs, sdk.StoreTypeIAVL, db)
	if err := mstore.LoadLatestVersion(); err != nil {
		t.Fatal(err)
	}

	input.paramsKeeper = params.NewKeeper(input.cdc, input.keyParams, input.tkeyParams, params.DefaultCodespace)

	input.accountKeeper = auth.NewAccountKeeper(
		input.cdc,
		input.keyAccount,
		input.paramsKeeper.Subspace(auth.DefaultParamspace),
		auth.ProtoBaseAccount,
	)

	input.bankKeeper = bank.NewBaseKeeper(
		input.accountKeeper,
		input.paramsKeeper.Subspace(bank.DefaultParamspace),
		bank.DefaultCodespace,
		input.ModuleAccountAddrs(),
	)

	input.supplyKeeper = supply.NewKeeper(input.cdc, input.keySupply, input.accountKeeper, input.bankKeeper, maccPerms)

	input.poaKeeper = poa.NewKeeper(input.keyPoa, input.cdc, input.paramsKeeper.Subspace(poa.DefaultParamspace))

	input.msRouter = core.NewRouter()
	input.msRouter.AddRoute(msgRouteNoop, func(ctx sdk.Context, msg core.MsMsg) sdk.Error {
		return nil
	})

	input.cdc.RegisterConcrete(TestMsg{}, "multisig/test-msg", nil)
	input.target = NewKeeper(input.keyMs, input.cdc, input.msRouter, input.paramsKeeper.Subspace(mstypes.DefaultParamspace))

	input.ctx = sdk.NewContext(mstore, abci.Header{ChainID: "test-chain-id"}, false, log.NewNopLogger())

	return input
}

func checkError(t *testing.T, expectedErr, receivedErr sdk.Error) {
	require.Equal(t, expectedErr.Codespace(), receivedErr.Codespace(), "Codespace")
	require.Equal(t, expectedErr.Code(), receivedErr.Code(), "code")
}

func TestKeeper_ExportGenesis(t *testing.T) {
	t.Parallel()

	input := setupTestInput(t)
	ctx := input.ctx
	target := input.target

	initGenesis := mstypes.GenesisState{
		Parameters: mstypes.DefaultParams(),
	}

	target.InitGenesis(ctx, initGenesis)
	exportGenesis := target.ExportGenesis(ctx)

	require.Equal(t, initGenesis, exportGenesis)
}

func TestKeeper_SubmitCallInvalidMsg(t *testing.T) {
	t.Parallel()

	input := setupTestInput(t)
	ctx := input.ctx
	target := input.target

	// check msg has no route condition
	{
		err := target.SubmitCall(ctx, NewTestMsg("", ""), "", sdk.AccAddress([]byte("addr1")))
		checkError(t, mstypes.ErrRouteDoesntExist(""), err)
	}
	// check msg has no type
	{
		err := target.SubmitCall(ctx, NewTestMsg(msgRouteNoop, ""), "", sdk.AccAddress([]byte("addr1")))
		checkError(t, mstypes.ErrEmptyType(0), err)
	}
}

func TestKeeper_SubmitCallUniqueness(t *testing.T) {
	t.Parallel()

	input := setupTestInput(t)
	ctx := input.ctx
	target := input.target

	testMsg, uniqueCallId := NewTestMsg(msgRouteNoop, "notEmpty"), "1"
	addr := sdk.AccAddress([]byte("addr1"))
	// add call
	{
		err := target.SubmitCall(ctx, testMsg, uniqueCallId, addr)
		require.Nil(t, err)
	}
	// confirm non-existing calId
	{
		err := target.Confirm(ctx, 1, addr)
		checkError(t, mstypes.ErrWrongCallId(0), err)
	}
	// get existing unique call
	{
		id, err := target.GetCallIDByUnique(ctx, uniqueCallId)
		require.Nil(t, err)
		require.Equal(t, uint64(0), id)
	}
	// get non-existing unique call
	{
		id, err := target.GetCallIDByUnique(ctx, "2")
		checkError(t, mstypes.ErrNotFoundUniqueID(""), err)
		require.Equal(t, uint64(0), id)
	}
	// check call with uniqueID already exists
	{
		err := target.SubmitCall(ctx, testMsg, uniqueCallId, addr)
		checkError(t, mstypes.ErrNotUniqueID(""), err)
	}
}

func TestModule_ValidateGenesis(t *testing.T) {
	t.Parallel()

	input := setupTestInput(t)
	app := NewAppModule(input.target, input.poaKeeper)

	genesis := mstypes.GenesisState{
		Parameters: mstypes.DefaultParams(),
	}

	// check OK
	require.Nil(t, app.ValidateGenesis(input.cdc.MustMarshalJSON(genesis)))

	// check minIntervalToExecute params error
	if mstypes.MinIntervalToExecute > 0 {
		genesis.Parameters.IntervalToExecute = mstypes.MinIntervalToExecute - 1
		require.NotNil(t, app.ValidateGenesis(input.cdc.MustMarshalJSON(genesis)))
	}
}
