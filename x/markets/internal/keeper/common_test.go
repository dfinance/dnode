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
	"github.com/dfinance/dvm-proto/go/vm_grpc"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"github.com/dfinance/dnode/x/common_vm"
	ccTypes "github.com/dfinance/dnode/x/currencies"
	"github.com/dfinance/dnode/x/markets/internal/types"
)

var (
	maccPerms map[string][]string = map[string][]string{
		auth.FeeCollectorName: nil,
	}
)

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
	keyAccount   *sdk.KVStoreKey
	keySupply    *sdk.KVStoreKey
	keyCC        *sdk.KVStoreKey
	keyVMStorage *sdk.KVStoreKey
	tKeyParams   *sdk.TransientStoreKey
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
	ccKeeper      ccTypes.Keeper
	paramsKeeper  params.Keeper
	keeper        Keeper
	//
	vmStorage common_vm.VMStorage
}

func (input *TestInput) ModuleAccountAddrs() map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range maccPerms {
		modAccAddrs[supply.NewModuleAddress(acc).String()] = true
	}

	return modAccAddrs
}

func NewTestInput(t *testing.T) TestInput {
	input := TestInput{
		cdc:          codec.New(),
		keyParams:    sdk.NewKVStoreKey(params.StoreKey),
		keyAccount:   sdk.NewKVStoreKey(auth.StoreKey),
		keySupply:    sdk.NewKVStoreKey(supply.StoreKey),
		keyCC:        sdk.NewKVStoreKey(ccTypes.StoreKey),
		keyVMStorage: sdk.NewKVStoreKey("key_vm_storage"),
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

	// init in-memory DB
	db := dbm.NewMemDB()
	mstore := store.NewCommitMultiStore(db)
	mstore.MountStoreWithDB(input.keyVMStorage, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.keyParams, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.keyAccount, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.keySupply, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.keyCC, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.tKeyParams, sdk.StoreTypeTransient, db)
	require.NoError(t, mstore.LoadLatestVersion(), "in-memory DB init")

	// create target and dependant keepers
	input.vmStorage = NewVMStorage(input.keyVMStorage)
	input.paramsKeeper = params.NewKeeper(input.cdc, input.keyParams, input.tKeyParams)
	input.accountKeeper = auth.NewAccountKeeper(input.cdc, input.keyAccount, input.paramsKeeper.Subspace(auth.DefaultParamspace), auth.ProtoBaseAccount)
	input.bankKeeper = bank.NewBaseKeeper(input.accountKeeper, input.paramsKeeper.Subspace(bank.DefaultParamspace), input.ModuleAccountAddrs())
	input.ccKeeper = ccTypes.NewKeeper(input.cdc, input.keyCC, input.paramsKeeper.Subspace(ccTypes.DefaultParamspace), input.bankKeeper, input.vmStorage)
	input.keeper = NewKeeper(input.cdc, input.paramsKeeper.Subspace(types.DefaultParamspace), input.ccKeeper)

	// create context
	input.ctx = sdk.NewContext(mstore, abci.Header{ChainID: "test-chain-id"}, false, log.NewNopLogger())

	// init genesis / params
	input.ccKeeper.InitDefaultGenesis(input.ctx)
	input.keeper.SetParams(input.ctx, types.DefaultParams())

	return input
}
