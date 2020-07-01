// +build unit

package keeper

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"github.com/dfinance/dnode/helpers/tests"
	"github.com/dfinance/dnode/x/cc_storage/internal/types"
	"github.com/dfinance/dnode/x/common_vm"
	"github.com/dfinance/dnode/x/vm"
)

type TestInput struct {
	cdc *codec.Codec
	ctx sdk.Context
	//
	keyCCStorage *sdk.KVStoreKey
	keyParams    *sdk.KVStoreKey
	tkeyParams   *sdk.TransientStoreKey
	keyVMS       *sdk.KVStoreKey
	//
	paramsKeeper params.Keeper
	keeper       Keeper
	//
	vmStorage common_vm.VMStorage
}

func NewTestInput(t *testing.T) TestInput {
	input := TestInput{
		cdc:          codec.New(),
		keyParams:    sdk.NewKVStoreKey(params.StoreKey),
		keyCCStorage: sdk.NewKVStoreKey(types.StoreKey),
		keyVMS:       sdk.NewKVStoreKey(vm.StoreKey),
		tkeyParams:   sdk.NewTransientStoreKey(params.TStoreKey),
	}

	// register codec
	sdk.RegisterCodec(input.cdc)
	codec.RegisterCrypto(input.cdc)

	// init in-memory DB
	db := dbm.NewMemDB()
	mstore := store.NewCommitMultiStore(db)
	mstore.MountStoreWithDB(input.keyParams, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.keyCCStorage, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.keyVMS, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.tkeyParams, sdk.StoreTypeTransient, db)
	err := mstore.LoadLatestVersion()
	if err != nil {
		t.Fatal(err)
	}

	// create test VM storage
	input.vmStorage = tests.NewVMStorage(input.keyVMS)

	// create target and dependant keepers
	input.paramsKeeper = params.NewKeeper(input.cdc, input.keyParams, input.tkeyParams)
	input.keeper = NewKeeper(input.cdc, input.keyCCStorage, input.paramsKeeper.Subspace(types.DefaultParamspace), input.vmStorage)

	// create context
	input.ctx = sdk.NewContext(mstore, abci.Header{ChainID: "test-chain-id"}, false, log.NewNopLogger())

	// init genesis
	input.keeper.InitDefaultGenesis(input.ctx)

	return input
}
