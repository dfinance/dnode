// +build unit

package keeper

import (
	"testing"
	"time"

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

	"github.com/dfinance/dnode/helpers/perms"
	"github.com/dfinance/dnode/helpers/tests"
	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/ccstorage"
	"github.com/dfinance/dnode/x/common_vm"
	"github.com/dfinance/dnode/x/markets"
	"github.com/dfinance/dnode/x/orders/internal/types"
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
	keyMarkets *sdk.KVStoreKey
	keyOrders  *sdk.KVStoreKey
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
	paramsKeeper  params.Keeper
	keeper        Keeper
	//
	vmStorage common_vm.VMStorage
}

func NewTestInput(t *testing.T, customMarketsPerms perms.Permissions) TestInput {
	input := TestInput{
		cdc:        codec.New(),
		keyParams:  sdk.NewKVStoreKey(params.StoreKey),
		keyAccount: sdk.NewKVStoreKey(auth.StoreKey),
		keySupply:  sdk.NewKVStoreKey(supply.StoreKey),
		keyCCS:     sdk.NewKVStoreKey(ccstorage.StoreKey),
		keyMarkets: sdk.NewKVStoreKey(markets.StoreKey),
		keyOrders:  sdk.NewKVStoreKey(types.StoreKey),
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
	types.RegisterCodec(input.cdc)

	// init in-memory DB
	db := dbm.NewMemDB()
	mstore := store.NewCommitMultiStore(db)
	mstore.MountStoreWithDB(input.keyVMS, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.keyParams, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.keyAccount, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.keySupply, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.keyCCS, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.keyMarkets, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.keyOrders, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.tKeyParams, sdk.StoreTypeTransient, db)
	require.NoError(t, mstore.LoadLatestVersion(), "in-memory DB init")

	// create custom markets module permission requester as some test do need to create a markets
	marketsRequester := types.RequestMarketsPerms()
	if customMarketsPerms != nil {
		marketsRequester = func() (moduleName string, modulePerms perms.Permissions) {
			moduleName, modulePerms = types.ModuleName, customMarketsPerms
			return
		}
	}

	// create target and dependant keepers
	input.vmStorage = tests.NewVMStorage(input.keyVMS)
	input.paramsKeeper = params.NewKeeper(input.cdc, input.keyParams, input.tKeyParams)
	input.accountKeeper = auth.NewAccountKeeper(input.cdc, input.keyAccount, input.paramsKeeper.Subspace(auth.DefaultParamspace), auth.ProtoBaseAccount)
	input.bankKeeper = bank.NewBaseKeeper(input.accountKeeper, input.paramsKeeper.Subspace(bank.DefaultParamspace), tests.ModuleAccountAddrs())
	input.supplyKeeper = supply.NewKeeper(input.cdc, input.keySupply, input.accountKeeper, input.bankKeeper, tests.MAccPerms)
	input.ccsKeeper = ccstorage.NewKeeper(
		input.cdc,
		input.keyCCS,
		input.vmStorage,
		markets.RequestCCStoragePerms(),
	)
	input.marketKeeper = markets.NewKeeper(
		input.cdc,
		input.keyMarkets,
		input.ccsKeeper,
		marketsRequester,
	)
	input.keeper = NewKeeper(input.cdc, input.keyOrders, input.bankKeeper, input.supplyKeeper, input.marketKeeper)

	// create context
	input.ctx = sdk.NewContext(mstore, abci.Header{ChainID: "test-chain-id"}, false, log.NewNopLogger())

	// init genesis / params
	input.ccsKeeper.InitDefaultGenesis(input.ctx)
	input.marketKeeper.InitDefaultGenesis(input.ctx)

	return input
}

func (i *TestInput) GetAccountBalance(address sdk.AccAddress, baseDenom string) (baseBalance, quoteBalance sdk.Int) {
	acc := i.accountKeeper.GetAccount(i.ctx, address)
	for _, coin := range acc.GetCoins() {
		if coin.Denom == baseDenom {
			baseBalance = coin.Amount
		}
		if coin.Denom == i.quoteDenom {
			quoteBalance = coin.Amount
		}
	}

	return
}

func NewBtcDfiMockOrder(direction types.Direction) types.Order {
	now := time.Now()

	return types.Order{
		ID:    dnTypes.NewIDFromUint64(0),
		Owner: sdk.AccAddress("wallet13jyjuz3kkdvqw"),
		Market: markets.MarketExtended{
			ID: dnTypes.NewIDFromUint64(0),
			BaseCurrency: ccstorage.Currency{
				Denom:    "btc",
				Decimals: 8,
			},
			QuoteCurrency: ccstorage.Currency{
				Denom:    "dfi",
				Decimals: 18,
			},
		},
		Direction: direction,
		Price:     sdk.NewUintFromString("1000000000000000000"),
		Quantity:  sdk.NewUintFromString("100000000"),
		Ttl:       60,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func NewEthDfiMockOrder(direction types.Direction) types.Order {
	now := time.Now()

	return types.Order{
		ID:    dnTypes.NewIDFromUint64(1),
		Owner: sdk.AccAddress("wallet13jyjuz3kkdvqx"),
		Market: markets.MarketExtended{
			ID: dnTypes.NewIDFromUint64(1),
			BaseCurrency: ccstorage.Currency{
				Denom:    "eth",
				Decimals: 18,
			},
			QuoteCurrency: ccstorage.Currency{
				Denom:    "dfi",
				Decimals: 18,
			},
		},
		Direction: direction,
		Price:     sdk.NewUintFromString("1000000000000000000"),
		Quantity:  sdk.NewUintFromString("1000000000000000000"),
		Ttl:       120,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func CompareOrders(t *testing.T, order1, order2 types.Order) {
	compareCurrency := func(logMsg string, c1, c2 ccstorage.Currency) {
		require.Equal(t, c1.Denom, c2.Denom, "%s: Denom", logMsg)
		require.Equal(t, c1.Decimals, c2.Decimals, "%s: Decimals", logMsg)
	}

	require.True(t, order1.ID.Equal(order2.ID), "ID")
	require.True(t, order1.Price.Equal(order2.Price), "Price")
	require.True(t, order1.Quantity.Equal(order2.Quantity), "Quantity")
	require.True(t, order1.Direction.Equal(order2.Direction), "Direction")
	require.Equal(t, order1.Owner.String(), order2.Owner.String(), "Owner")
	require.Equal(t, order1.Ttl, order2.Ttl, "Ttl")
	require.True(t, order1.CreatedAt.Equal(order2.CreatedAt), "CreatedAt")
	require.True(t, order1.UpdatedAt.Equal(order2.UpdatedAt), "UpdatedAt")
	//
	require.True(t, order1.Market.ID.Equal(order2.Market.ID), "Market.ID")
	compareCurrency("baseAsset", order1.Market.BaseCurrency, order2.Market.BaseCurrency)
	compareCurrency("quoteAsset", order1.Market.QuoteCurrency, order2.Market.QuoteCurrency)
}
