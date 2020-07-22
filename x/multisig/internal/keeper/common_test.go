// +build unit

package keeper

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/supply"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"github.com/dfinance/dnode/helpers/tests"
	"github.com/dfinance/dnode/x/core/msmodule"
	"github.com/dfinance/dnode/x/multisig/internal/types"
)

const (
	MockMsgType     = "msType"
	MockMsgRouteOk  = "mockRouteOk"
	MockMsgRouteErr = "mockRouteErr"
)

type TestInput struct {
	cdc *codec.Codec
	ctx sdk.Context
	//
	keyParams  *sdk.KVStoreKey
	keyAccount *sdk.KVStoreKey
	keySupply  *sdk.KVStoreKey
	keyMS      *sdk.KVStoreKey
	tkeyParams *sdk.TransientStoreKey
	//
	msRouter msmodule.MsRouter
	//
	accountKeeper auth.AccountKeeper
	bankKeeper    bank.Keeper
	supplyKeeper  supply.Keeper
	paramsKeeper  params.Keeper
	target        Keeper
}

func NewTestInput(t *testing.T) TestInput {
	input := TestInput{
		cdc:        codec.New(),
		keyParams:  sdk.NewKVStoreKey(params.StoreKey),
		keyAccount: sdk.NewKVStoreKey(auth.StoreKey),
		keySupply:  sdk.NewKVStoreKey(supply.StoreKey),
		keyMS:      sdk.NewKVStoreKey(types.StoreKey),
		tkeyParams: sdk.NewTransientStoreKey(params.TStoreKey),
	}

	// register codec
	types.RegisterCodec(input.cdc)
	sdk.RegisterCodec(input.cdc)
	codec.RegisterCrypto(input.cdc)

	// init in-memory DB
	db := dbm.NewMemDB()
	mstore := store.NewCommitMultiStore(db)
	mstore.MountStoreWithDB(input.keyAccount, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.keySupply, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.keyParams, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.tkeyParams, sdk.StoreTypeTransient, db)
	mstore.MountStoreWithDB(input.keyMS, sdk.StoreTypeIAVL, db)
	if err := mstore.LoadLatestVersion(); err != nil {
		t.Fatal(err)
	}

	// create target and dependant keepers
	input.paramsKeeper = params.NewKeeper(input.cdc, input.keyParams, input.tkeyParams)
	input.accountKeeper = auth.NewAccountKeeper(input.cdc, input.keyAccount, input.paramsKeeper.Subspace(auth.DefaultParamspace), auth.ProtoBaseAccount)
	input.bankKeeper = bank.NewBaseKeeper(input.accountKeeper, input.paramsKeeper.Subspace(bank.DefaultParamspace), tests.ModuleAccountAddrs())
	input.supplyKeeper = supply.NewKeeper(input.cdc, input.keySupply, input.accountKeeper, input.bankKeeper, tests.MAccPerms)

	// init multisig router
	input.msRouter = msmodule.NewMsRouter()
	input.msRouter.AddRoute(MockMsgRouteOk, func(ctx sdk.Context, msg msmodule.MsMsg) error { return nil })
	input.msRouter.AddRoute(MockMsgRouteErr, func(ctx sdk.Context, msg msmodule.MsMsg) error { return fmt.Errorf("error") })

	// register mock multisig msg
	input.cdc.RegisterConcrete(MockMsMsg{}, "multisig/mock-msg", nil)

	// create target keeper
	input.target = NewKeeper(input.cdc, input.keyMS, input.paramsKeeper.Subspace(types.DefaultParamspace), input.msRouter)

	// create context
	input.ctx = sdk.NewContext(mstore, abci.Header{ChainID: "test-chain-id"}, false, log.NewNopLogger())

	return input
}

type MockMsMsg struct {
	isValid bool
}

func (m MockMsMsg) Route() string { return MockMsgRouteOk }
func (m MockMsMsg) Type() string  { return MockMsgType }
func (m MockMsMsg) ValidateBasic() error {
	if !m.isValid {
		return fmt.Errorf("some error")
	}
	return nil
}
func NewMockMsMsg(isValid bool) MockMsMsg {
	return MockMsMsg{isValid: isValid}
}

type CustomMockMsMsg struct {
	msType  string
	msRoute string
}

func (m CustomMockMsMsg) Route() string        { return m.msRoute }
func (m CustomMockMsMsg) Type() string         { return m.msType }
func (m CustomMockMsMsg) ValidateBasic() error { return nil }
