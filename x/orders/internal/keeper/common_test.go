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
	"github.com/dfinance/dvm-proto/go/vm_grpc"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/common_vm"
	"github.com/dfinance/dnode/x/currencies_register"
	"github.com/dfinance/dnode/x/markets"
	"github.com/dfinance/dnode/x/orders/internal/types"
)

var (
	// BankKeeper, SupplyKeeper dependency
	maccPerms = map[string][]string{
		types.ModuleName: {supply.Burner},
	}
)

// BankKeeper, SupplyKeeper dependency.
func ModuleAccountAddrs() map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range maccPerms {
		modAccAddrs[supply.NewModuleAddress(acc).String()] = true
	}

	return modAccAddrs
}

// Mock VM storage implementation.
type VMStorage struct {
	storeKey sdk.StoreKey
}

func NewVMStorage(storeKey sdk.StoreKey) VMStorage {
	return VMStorage{
		storeKey: storeKey,
	}
}

func (storage VMStorage) GetOracleAccessPath(_ string) *vm_grpc.VMAccessPath {
	return &vm_grpc.VMAccessPath{}
}

func (storage VMStorage) SetValue(ctx sdk.Context, accessPath *vm_grpc.VMAccessPath, value []byte) {
	store := ctx.KVStore(storage.storeKey)
	store.Set(common_vm.MakePathKey(accessPath), value)
}

func (storage VMStorage) GetValue(ctx sdk.Context, accessPath *vm_grpc.VMAccessPath) []byte {
	store := ctx.KVStore(storage.storeKey)
	return store.Get(common_vm.MakePathKey(accessPath))
}

func (storage VMStorage) DelValue(ctx sdk.Context, accessPath *vm_grpc.VMAccessPath) {
	store := ctx.KVStore(storage.storeKey)
	store.Delete(common_vm.MakePathKey(accessPath))
}

func (storage VMStorage) HasValue(ctx sdk.Context, accessPath *vm_grpc.VMAccessPath) bool {
	store := ctx.KVStore(storage.storeKey)
	return store.Has(common_vm.MakePathKey(accessPath))
}

// Module keeper tests input.
type TestInput struct {
	cdc *codec.Codec
	ctx sdk.Context
	//
	keyParams    *sdk.KVStoreKey
	keyCR        *sdk.KVStoreKey
	keyVMStorage *sdk.KVStoreKey
	keyAuth      *sdk.KVStoreKey
	keySupply    *sdk.KVStoreKey
	keyOrders    *sdk.KVStoreKey
	tKeyParams   *sdk.TransientStoreKey
	//
	baseBtcDenom    string
	baseBtcDecimals uint8
	baseEthDenom    string
	baseEthDecimals uint8
	quoteDenom      string
	quoteDecimals   uint8
	//
	vmStorage     common_vm.VMStorage
	paramsKeeper  params.Keeper
	crKeeper      currencies_register.Keeper
	marketKeeper  markets.Keeper
	accountKeeper auth.AccountKeeper
	bankKeeper    bank.Keeper
	supplyKeeper  supply.Keeper
	keeper        Keeper
}

func NewTestInput(t *testing.T) TestInput {
	input := TestInput{
		cdc:          codec.New(),
		keyParams:    sdk.NewKVStoreKey("key_params"),
		keyCR:        sdk.NewKVStoreKey("key_cr"),
		keyVMStorage: sdk.NewKVStoreKey("key_vm_storage"),
		keyAuth:      sdk.NewKVStoreKey("key_auth"),
		keySupply:    sdk.NewKVStoreKey("key_supply"),
		keyOrders:    sdk.NewKVStoreKey("key_orders"),
		tKeyParams:   sdk.NewTransientStoreKey("tkey_params"),
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
	mstore.MountStoreWithDB(input.keyParams, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.keyCR, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.keyVMStorage, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.keyAuth, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.keySupply, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.keyOrders, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.tKeyParams, sdk.StoreTypeTransient, db)
	require.NoError(t, mstore.LoadLatestVersion(), "in-memory DB init")

	// create target and dependant keepers
	input.vmStorage = NewVMStorage(input.keyVMStorage)
	input.paramsKeeper = params.NewKeeper(input.cdc, input.keyParams, input.tKeyParams)
	input.crKeeper = currencies_register.NewKeeper(input.cdc, input.keyCR, input.vmStorage)
	input.marketKeeper = markets.NewKeeper(input.cdc, input.paramsKeeper.Subspace(markets.DefaultParamspace), input.crKeeper)
	input.accountKeeper = auth.NewAccountKeeper(input.cdc, input.keyAuth, input.paramsKeeper.Subspace(auth.DefaultParamspace), auth.ProtoBaseAccount)
	input.bankKeeper = bank.NewBaseKeeper(input.accountKeeper, input.paramsKeeper.Subspace(bank.DefaultParamspace), ModuleAccountAddrs())
	input.supplyKeeper = supply.NewKeeper(input.cdc, input.keySupply, input.accountKeeper, input.bankKeeper, maccPerms)
	input.keeper = NewKeeper(input.keyOrders, input.cdc, input.bankKeeper, input.supplyKeeper, input.marketKeeper)

	// create context
	input.ctx = sdk.NewContext(mstore, abci.Header{ChainID: "test-chain-id"}, false, log.NewNopLogger())

	// init params
	input.marketKeeper.SetParams(input.ctx, markets.DefaultParams())

	// init currencies
	baseSupply, ok := sdk.NewIntFromString("100000000000000")
	require.True(t, ok)
	quoteSupply, ok := sdk.NewIntFromString("100000000000000000000000000")
	require.True(t, ok)

	ownerAddr := make([]byte, common_vm.VMAddressLength)

	err := input.crKeeper.AddCurrencyInfo(
		input.ctx,
		input.baseBtcDenom,
		input.baseBtcDecimals,
		false,
		ownerAddr,
		baseSupply,
		[]byte("01fe7c965b1c008c5974c7750959fa10189e803225d5057207563553922a09f906"))
	require.NoError(t, err)

	err = input.crKeeper.AddCurrencyInfo(
		input.ctx,
		input.baseEthDenom,
		input.baseEthDecimals,
		false,
		ownerAddr,
		baseSupply,
		[]byte("01f8799f504905a182aff8d5fc102da1d73b8bec199147bb5512af6e99006baeb6"))
	require.NoError(t, err)

	err = input.crKeeper.AddCurrencyInfo(
		input.ctx,
		input.quoteDenom,
		input.quoteDecimals,
		false,
		ownerAddr,
		quoteSupply,
		[]byte("018bfc024222e94fbed60ff0c9c1cf48c5b2809d83c82f513b2c385e21ba8a2d35"))
	require.NoError(t, err)

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
		Owner: sdk.AccAddress("wallet13jyjuz3kkdvqw8u4qfkwd94emdl3vx394kn07h"),
		Market: markets.MarketExtended{
			ID: dnTypes.NewIDFromUint64(0),
			BaseCurrency: currencies_register.CurrencyInfo{
				Denom:    []byte("btc"),
				Decimals: 8,
			},
			QuoteCurrency: currencies_register.CurrencyInfo{
				Denom:    []byte("dfi"),
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
		Owner: sdk.AccAddress("wallet13jyjuz3kkdvqw8u4qfkwd94emdl3vx394kn07i"),
		Market: markets.MarketExtended{
			ID: dnTypes.NewIDFromUint64(1),
			BaseCurrency: currencies_register.CurrencyInfo{
				Denom:    []byte("eth"),
				Decimals: 18,
			},
			QuoteCurrency: currencies_register.CurrencyInfo{
				Denom:    []byte("dfi"),
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
	compareCurrency := func(logMsg string, c1, c2 currencies_register.CurrencyInfo) {
		require.Equal(t, c1.Owner, c2.Owner, "%s: Owner", logMsg)
		require.Equal(t, c1.Denom, c2.Denom, "%s: Denom", logMsg)
		require.Equal(t, c1.Decimals, c2.Decimals, "%s: Decimals", logMsg)
		//require.Equal(t, c1.TotalSupply.String(), c2.TotalSupply.String(), "%s: TotalSupply", logMsg)
		require.Equal(t, c1.IsToken, c2.IsToken, "%s: IsToken", logMsg)
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
