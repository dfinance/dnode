// +build unit

package keeper

import (
	"math/rand"
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
	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/ccstorage"
	"github.com/dfinance/dnode/x/common_vm"
	"github.com/dfinance/dnode/x/markets"
	"github.com/dfinance/dnode/x/orderbook/internal/types"
	"github.com/dfinance/dnode/x/orders"
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
	keyOrders  *sdk.KVStoreKey
	keyOB      *sdk.KVStoreKey
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
	ccsKeeper     ccstorage.Keeper
	marketKeeper  markets.Keeper
	orderKeeper   orders.Keeper
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
		keyOrders:  sdk.NewKVStoreKey(orders.StoreKey),
		keyOB:      sdk.NewKVStoreKey(types.StoreKey),
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
	markets.RegisterCodec(input.cdc)
	auth.RegisterCodec(input.cdc)
	bank.RegisterCodec(input.cdc)
	supply.RegisterCodec(input.cdc)
	orders.RegisterCodec(input.cdc)
	types.RegisterCodec(input.cdc)

	// init in-memory DB
	db := dbm.NewMemDB()
	mstore := store.NewCommitMultiStore(db)
	mstore.MountStoreWithDB(input.keyVMS, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.keyParams, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.keyAccount, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.keySupply, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.keyCCS, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.keyOrders, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.keyOB, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.tKeyParams, sdk.StoreTypeTransient, db)
	require.NoError(t, mstore.LoadLatestVersion(), "in-memory DB init")

	// create target and dependant keepers
	input.vmStorage = tests.NewVMStorage(input.keyVMS)
	input.paramsKeeper = params.NewKeeper(input.cdc, input.keyParams, input.tKeyParams)
	input.accountKeeper = auth.NewAccountKeeper(input.cdc, input.keyAccount, input.paramsKeeper.Subspace(auth.DefaultParamspace), auth.ProtoBaseAccount)
	input.bankKeeper = bank.NewBaseKeeper(input.accountKeeper, input.paramsKeeper.Subspace(bank.DefaultParamspace), tests.ModuleAccountAddrs())
	input.supplyKeeper = supply.NewKeeper(input.cdc, input.keySupply, input.accountKeeper, input.bankKeeper, tests.MAccPerms)
	input.ccsKeeper = ccstorage.NewKeeper(
		input.cdc,
		input.keyCCS,
		input.paramsKeeper.Subspace(ccstorage.DefaultParamspace),
		input.vmStorage,
		markets.RequestCCStoragePerms(),
	)
	input.marketKeeper = markets.NewKeeper(
		input.cdc,
		input.paramsKeeper.Subspace(markets.DefaultParamspace),
		input.ccsKeeper,
		orders.RequestMarketsPerms(),
	)
	input.orderKeeper = orders.NewKeeper(
		input.cdc,
		input.keyOrders,
		input.bankKeeper,
		input.supplyKeeper,
		input.marketKeeper,
		types.RequestOrdersPerms(),
	)
	input.keeper = NewKeeper(input.keyOB, input.cdc, input.orderKeeper)

	// create context
	input.ctx = sdk.NewContext(mstore, abci.Header{ChainID: "test-chain-id"}, false, log.NewNopLogger())

	// init genesis / params
	input.ccsKeeper.InitDefaultGenesis(input.ctx)
	input.marketKeeper.InitDefaultGenesis(input.ctx)

	return input
}

func NewMockHistoryItem(marketID dnTypes.ID, blockHeight int64) types.HistoryItem {
	return types.HistoryItem{
		MarketID:         marketID,
		ClearancePrice:   sdk.NewUint(rand.Uint64()),
		BidOrdersCount:   rand.Int(),
		AskOrdersCount:   rand.Int(),
		BidVolume:        sdk.NewUint(rand.Uint64()),
		AskVolume:        sdk.NewUint(rand.Uint64()),
		MatchedBidVolume: sdk.NewUint(rand.Uint64()),
		MatchedAskVolume: sdk.NewUint(rand.Uint64()),
		Timestamp:        rand.Int63(),
		BlockHeight:      blockHeight,
	}
}
