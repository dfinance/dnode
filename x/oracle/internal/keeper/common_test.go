// +build unit

package keeper

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/mock"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/supply"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"github.com/dfinance/dnode/helpers/tests"
	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/common_vm"
	"github.com/dfinance/dnode/x/oracle/internal/types"
	"github.com/dfinance/dnode/x/poa"
	"github.com/dfinance/dnode/x/vm"
)

type TestInput struct {
	cdc *codec.Codec
	ctx sdk.Context

	keyParams  *sdk.KVStoreKey
	keyAccount *sdk.KVStoreKey
	keySupply  *sdk.KVStoreKey
	keyPOA     *sdk.KVStoreKey
	keyOracle  *sdk.KVStoreKey
	keyVMS     *sdk.KVStoreKey
	tKeyParams *sdk.TransientStoreKey

	accountKeeper auth.AccountKeeper
	bankKeeper    bank.Keeper
	supplyKeeper  supply.Keeper
	paramsKeeper  params.Keeper
	poaKeeper     poa.Keeper
	vmStorage     common_vm.VMStorage
	keeper        Keeper

	addresses    []sdk.AccAddress
	stdAssetCode dnTypes.AssetCode
	stdAssets    types.Assets
	stdNominee   string
}

// NewTestInput returned mocked object for testing
func NewTestInput(t *testing.T) TestInput {
	input := TestInput{
		cdc:        codec.New(),
		keyParams:  sdk.NewKVStoreKey(params.StoreKey),
		keyAccount: sdk.NewKVStoreKey(auth.StoreKey),
		keySupply:  sdk.NewKVStoreKey(supply.StoreKey),
		keyVMS:     sdk.NewKVStoreKey(vm.StoreKey),
		keyOracle:  sdk.NewKVStoreKey(types.StoreKey),
		tKeyParams: sdk.NewTransientStoreKey(params.TStoreKey),
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
	mstore.MountStoreWithDB(input.keyOracle, sdk.StoreTypeIAVL, db)
	mstore.MountStoreWithDB(input.tKeyParams, sdk.StoreTypeTransient, db)
	require.NoError(t, mstore.LoadLatestVersion(), "in-memory DB init")

	// create target and dependant keepers
	input.vmStorage = tests.NewVMStorage(input.keyVMS)
	input.paramsKeeper = params.NewKeeper(input.cdc, input.keyParams, input.tKeyParams)
	input.accountKeeper = auth.NewAccountKeeper(input.cdc, input.keyAccount, input.paramsKeeper.Subspace(auth.DefaultParamspace), auth.ProtoBaseAccount)
	input.bankKeeper = bank.NewBaseKeeper(input.accountKeeper, input.paramsKeeper.Subspace(bank.DefaultParamspace), tests.ModuleAccountAddrs())
	input.keeper = NewKeeper(input.keyOracle, input.cdc, input.paramsKeeper.Subspace(types.DefaultParamspace), input.vmStorage)

	// create context
	input.ctx = sdk.NewContext(mstore, abci.Header{ChainID: "test-chain-id"}, false, log.NewNopLogger())

	// init genesis / params
	input.keeper.SetParams(input.ctx, types.DefaultParams())

	valTokens := sdk.TokensFromConsensusPower(50)

	accountsQuantity := 10
	_, input.addresses, _, _ = mock.CreateGenAccounts(accountsQuantity,
		sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, valTokens)))

	input.stdNominee = input.addresses[accountsQuantity-1].String()

	input.stdAssetCode = dnTypes.AssetCode("btc_dfi")

	input.stdAssets = types.Assets{types.NewAsset(input.stdAssetCode, []types.Oracle{}, true)}

	params := types.Params{
		Assets:   types.Assets{types.NewAsset(input.stdAssetCode, []types.Oracle{}, true)},
		Nominees: []string{input.stdNominee},
		PostPrice: types.PostPriceParams{
			ReceivedAtDiffInS: 60 * 60,
		},
	}

	input.keeper.SetParams(input.ctx, params)

	return input
}
