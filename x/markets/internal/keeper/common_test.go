// +build unit

package keeper

import (
	"testing"

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

	"github.com/dfinance/dnode/helpers/tests"
	"github.com/dfinance/dnode/x/ccstorage"
	"github.com/dfinance/dnode/x/common_vm"
	"github.com/dfinance/dnode/x/markets/internal/types"
	"github.com/dfinance/dnode/x/vm"
)

// Module keeper tests input.
type TestInput struct {
	cdc *codec.Codec
	ctx sdk.Context
	//
	keyParams  *sdk.KVStoreKey
	keyAccount *sdk.KVStoreKey
	keySupply  *sdk.KVStoreKey
	keyCCS     *sdk.KVStoreKey
	keyVMS     *sdk.KVStoreKey
	tKeyParams *sdk.TransientStoreKey
	//
	baseBtcDenom    string
	baseBtcDecimals uint8
	baseEthDenom    string
	baseEthDecimals uint8
	quoteDenom      string
	quoteDecimals   uint8
	//
	accountKeeper auth.AccountKeeper
	bankKeeper    bank.Keeper
	supplyKeeper  supply.Keeper
	ccsStorage    ccstorage.Keeper
	paramsKeeper  params.Keeper
	keeper        Keeper
	//
	vmStorage common_vm.VMStorage
}

func NewTestInput(t *testing.T) TestInput {
	input := TestInput{
		cdc:        codec.New(),
		keyParams:  sdk.NewKVStoreKey(params.StoreKey),
		keyAccount: sdk.NewKVStoreKey(auth.StoreKey),
		keySupply:  sdk.NewKVStoreKey(supply.StoreKey),
		keyCCS:     sdk.NewKVStoreKey(ccstorage.StoreKey),
		keyVMS:     sdk.NewKVStoreKey(vm.StoreKey),
		tKeyParams: sdk.NewTransientStoreKey(params.TStoreKey),
		//
		baseBtcDenom:    "btc",
		baseBtcDecimals: 8,
		baseEthDenom:    "eth",
		baseEthDecimals: 18,
		quoteDenom:      "dfi",
		quoteDecimals:   18,
	}

	// register codec
	sdk.RegisterCodec(input.cdc)
	codec.RegisterCrypto(input.cdc)

	// init in-memory DB
	db := dbm.NewMemDB()
	mstore := store.NewCommitMultiStore(db)
	mstore.MountStoreWithDB(input.keyVMS, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.keyParams, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.keyAccount, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.keySupply, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.keyCCS, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.tKeyParams, sdk.StoreTypeTransient, db)
	require.NoError(t, mstore.LoadLatestVersion(), "in-memory DB init")

	// create target and dependant keepers
	input.vmStorage = tests.NewVMStorage(input.keyVMS)
	input.paramsKeeper = params.NewKeeper(input.cdc, input.keyParams, input.tKeyParams)
	input.accountKeeper = auth.NewAccountKeeper(input.cdc, input.keyAccount, input.paramsKeeper.Subspace(auth.DefaultParamspace), auth.ProtoBaseAccount)
	input.bankKeeper = bank.NewBaseKeeper(input.accountKeeper, input.paramsKeeper.Subspace(bank.DefaultParamspace), tests.ModuleAccountAddrs())
	input.ccsStorage = ccstorage.NewKeeper(input.cdc, input.keyCCS, input.paramsKeeper.Subspace(ccstorage.DefaultParamspace), input.vmStorage)
	input.keeper = NewKeeper(input.cdc, input.paramsKeeper.Subspace(types.DefaultParamspace), input.ccsStorage)

	// create context
	input.ctx = sdk.NewContext(mstore, abci.Header{ChainID: "test-chain-id"}, false, log.NewNopLogger())

	// init genesis / params
	input.ccsStorage.InitDefaultGenesis(input.ctx)
	input.keeper.SetParams(input.ctx, types.DefaultParams())

	return input
}
