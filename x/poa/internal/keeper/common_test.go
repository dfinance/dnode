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
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/supply"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"github.com/dfinance/dnode/helpers/tests"
	"github.com/dfinance/dnode/x/poa/internal/types"
)

const (
	ethAddress1       = "0x29D7d1dd5B6f9C864d9db560D72a247c178aE86B"
	ethAddress2       = "0x29D7d1dd5B6f9C864d9db560D72a247c178aE87B"
	ethAddress3       = "0x29D7d1dd5B6f9C864d9db560D72a247c178aE88B"
	ethAddress4       = "0x29D7d1dd5B6f9C864d9db560D72a247c178aE89B"
	ethAddressInvalid = "0x82A978B3f5962A5b0957d9ee9eEf472EE55B42F"
)

var (
	sdkAddress1 = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	sdkAddress2 = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	sdkAddress3 = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	sdkAddress4 = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
)

type TestInput struct {
	cdc *codec.Codec
	ctx sdk.Context
	//
	keyAccount *sdk.KVStoreKey
	keySupply  *sdk.KVStoreKey
	keyPOA     *sdk.KVStoreKey
	keyParams  *sdk.KVStoreKey
	tkeyParams *sdk.TransientStoreKey
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
		keyAccount: sdk.NewKVStoreKey(auth.StoreKey),
		keySupply:  sdk.NewKVStoreKey(supply.StoreKey),
		keyParams:  sdk.NewKVStoreKey(params.StoreKey),
		keyPOA:     sdk.NewKVStoreKey(types.StoreKey),
		tkeyParams: sdk.NewTransientStoreKey(params.TStoreKey),
	}

	// register codec
	types.RegisterCodec(input.cdc)
	auth.RegisterCodec(input.cdc)
	bank.RegisterCodec(input.cdc)
	staking.RegisterCodec(input.cdc)
	sdk.RegisterCodec(input.cdc)
	codec.RegisterCrypto(input.cdc)

	// init in-memory DB
	db := dbm.NewMemDB()
	mstore := store.NewCommitMultiStore(db)
	mstore.MountStoreWithDB(input.keyAccount, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.keySupply, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.keyParams, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.keyPOA, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.tkeyParams, sdk.StoreTypeTransient, db)
	if err := mstore.LoadLatestVersion(); err != nil {
		t.Fatal(err)
	}

	// create target and dependant keepers
	input.paramsKeeper = params.NewKeeper(input.cdc, input.keyParams, input.tkeyParams)
	input.accountKeeper = auth.NewAccountKeeper(input.cdc, input.keyAccount, input.paramsKeeper.Subspace(auth.DefaultParamspace), auth.ProtoBaseAccount)
	input.bankKeeper = bank.NewBaseKeeper(input.accountKeeper, input.paramsKeeper.Subspace(bank.DefaultParamspace), tests.ModuleAccountAddrs())
	input.supplyKeeper = supply.NewKeeper(input.cdc, input.keySupply, input.accountKeeper, input.bankKeeper, tests.MAccPerms)

	// create target keeper
	input.target = NewKeeper(input.cdc, input.keyPOA, input.paramsKeeper.Subspace(types.DefaultParamspace))

	// create context
	input.ctx = sdk.NewContext(mstore, abci.Header{ChainID: "test-chain-id"}, false, log.NewNopLogger())

	// init genesis
	input.target.InitDefaultGenesis(input.ctx)

	return input
}
