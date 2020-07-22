// +build unit

package keeper

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"github.com/dfinance/dnode/helpers/tests"
	"github.com/dfinance/dnode/x/ccstorage"
	"github.com/dfinance/dnode/x/common_vm"
	"github.com/dfinance/dnode/x/vm"
	"github.com/dfinance/dnode/x/vmauth/internal/types"
)

type TestInput struct {
	cdc *codec.Codec
	ctx sdk.Context
	//
	keyParams  *sdk.KVStoreKey
	keyVMS     *sdk.KVStoreKey
	keyCCS     *sdk.KVStoreKey
	keyAccount *sdk.KVStoreKey
	tkeyParams *sdk.TransientStoreKey
	//
	paramsKeeper  params.Keeper
	bankKeeper    bank.BaseKeeper
	ccsStorage    ccstorage.Keeper
	accountKeeper VMAccountKeeper
	//
	vmStorage common_vm.VMStorage
}

// Create account without keeper involvement.
func (input *TestInput) CreateAccount(t *testing.T, coins sdk.Coins) exported.Account {
	if coins == nil {
		coins = sdk.NewCoins()
	}

	addr := secp256k1.GenPrivKey().PubKey().Address().Bytes()
	acc := input.accountKeeper.NewAccountWithAddress(input.ctx, addr)

	if len(coins) > 0 {
		require.NoError(t, acc.SetCoins(coins), "creating acc: setting coins")
	}

	return acc
}

func NewTestInput(t *testing.T) TestInput {
	input := TestInput{
		cdc:        codec.New(),
		keyParams:  sdk.NewKVStoreKey(params.StoreKey),
		keyVMS:     sdk.NewKVStoreKey(vm.StoreKey),
		keyCCS:     sdk.NewKVStoreKey(ccstorage.StoreKey),
		keyAccount: sdk.NewKVStoreKey(auth.StoreKey),
		tkeyParams: sdk.NewTransientStoreKey(params.TStoreKey),
	}

	// register codec
	auth.RegisterCodec(input.cdc)
	sdk.RegisterCodec(input.cdc)
	codec.RegisterCrypto(input.cdc)

	// init in-memory DB
	db := dbm.NewMemDB()
	mstore := store.NewCommitMultiStore(db)
	mstore.MountStoreWithDB(input.keyParams, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.keyVMS, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.keyCCS, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.keyAccount, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.tkeyParams, sdk.StoreTypeTransient, db)
	err := mstore.LoadLatestVersion()
	if err != nil {
		t.Fatal(err)
	}

	// create test VM storage
	input.vmStorage = tests.NewVMStorage(input.keyVMS)

	// create target and dependant keepers
	input.paramsKeeper = params.NewKeeper(input.cdc, input.keyParams, input.tkeyParams)
	input.ccsStorage = ccstorage.NewKeeper(
		input.cdc,
		input.keyCCS,
		input.vmStorage,
		types.RequestCCStoragePerms(),
	)
	input.accountKeeper = NewKeeper(input.cdc, input.keyAccount, input.paramsKeeper.Subspace(auth.DefaultParamspace), input.ccsStorage, auth.ProtoBaseAccount)
	input.bankKeeper = bank.NewBaseKeeper(input.accountKeeper, input.paramsKeeper.Subspace(bank.DefaultParamspace), make(map[string]bool))

	// create context
	input.ctx = sdk.NewContext(mstore, abci.Header{ChainID: "test-chain-id"}, false, log.NewNopLogger())

	// init genesis
	input.ccsStorage.InitDefaultGenesis(input.ctx)

	return input
}
