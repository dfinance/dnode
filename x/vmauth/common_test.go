package vmauth

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/dfinance/dvm-proto/go/vm_grpc"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"github.com/dfinance/dnode/x/common_vm"
)

// VM storage.
type VMStorageImpl struct {
	storeKey sdk.StoreKey
}

// Test input.
type testInput struct {
	cdc *codec.Codec
	ctx sdk.Context

	keyMain      *sdk.KVStoreKey
	keyAccount   *sdk.KVStoreKey
	keyParams    *sdk.KVStoreKey
	keyVMStorage *sdk.KVStoreKey
	tkeyParams   *sdk.TransientStoreKey

	paramsKeeper  params.Keeper
	accountKeeper VMAccountKeeper
	vmStorage     common_vm.VMStorage
}

// Create VM storage for tests.

func NewVMStorage(storeKey sdk.StoreKey) VMStorageImpl {
	return VMStorageImpl{
		storeKey: storeKey,
	}
}

func (storage VMStorageImpl) GetOracleAccessPath(_ string) *vm_grpc.VMAccessPath {
	return &vm_grpc.VMAccessPath{}
}

func (storage VMStorageImpl) SetValue(ctx sdk.Context, accessPath *vm_grpc.VMAccessPath, value []byte) {
	store := ctx.KVStore(storage.storeKey)
	store.Set(common_vm.MakePathKey(*accessPath), value)
}

func (storage VMStorageImpl) GetValue(ctx sdk.Context, accessPath *vm_grpc.VMAccessPath) []byte {
	store := ctx.KVStore(storage.storeKey)
	return store.Get(common_vm.MakePathKey(*accessPath))
}

func (storage VMStorageImpl) DelValue(ctx sdk.Context, accessPath *vm_grpc.VMAccessPath) {
	store := ctx.KVStore(storage.storeKey)
	store.Delete(common_vm.MakePathKey(*accessPath))
}

func newTestInput(t *testing.T) testInput {
	input := testInput{
		cdc:          codec.New(),
		keyMain:      sdk.NewKVStoreKey("main"),
		keyAccount:   sdk.NewKVStoreKey("acc"),
		keyVMStorage: sdk.NewKVStoreKey("vm_storage"),
	}

	auth.RegisterCodec(input.cdc)
	sdk.RegisterCodec(input.cdc)
	codec.RegisterCrypto(input.cdc)

	db := dbm.NewMemDB()
	mstore := store.NewCommitMultiStore(db)
	mstore.MountStoreWithDB(input.keyMain, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.keyAccount, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.keyVMStorage, sdk.StoreTypeIAVL, db)

	err := mstore.LoadLatestVersion()
	if err != nil {
		t.Fatal(err)
	}

	// The ParamsKeeper handles parameter storage for the application
	input.paramsKeeper = params.NewKeeper(input.cdc, input.keyParams, input.tkeyParams, params.DefaultCodespace)

	// Init vm storage.
	input.vmStorage = NewVMStorage(input.keyVMStorage)

	// The AccountKeeper handles address -> account lookups
	input.accountKeeper = NewVMAccountKeeper(
		input.cdc,
		input.keyAccount,
		input.paramsKeeper.Subspace(auth.DefaultParamspace),
		input.vmStorage,
		auth.ProtoBaseAccount,
	)

	// Setup context.
	input.ctx = sdk.NewContext(mstore, abci.Header{ChainID: "test-chain-id"}, false, log.NewNopLogger())

	return input
}
