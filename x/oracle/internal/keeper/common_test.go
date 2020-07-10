// +build unit

package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/supply"
	"github.com/dfinance/dnode/helpers/tests"
	"github.com/dfinance/dnode/x/common_vm"
	"github.com/dfinance/dnode/x/poa"
	"github.com/dfinance/dnode/x/vm"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/mock"
	"github.com/dfinance/dvm-proto/go/vm_grpc"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"

	"github.com/dfinance/dnode/x/oracle/internal/types"
)

type testHelper struct {
	mApp     *mock.App
	keeper   Keeper
	addrs    []sdk.AccAddress
	pubKeys  []crypto.PubKey
	privKeys []crypto.PrivKey
}

type VMStorageImpl struct {
}

func NewVMStorage() VMStorageImpl {
	return VMStorageImpl{}
}

func (storage VMStorageImpl) GetOracleAccessPath(_ string) *vm_grpc.VMAccessPath {
	return &vm_grpc.VMAccessPath{}
}

func (storage VMStorageImpl) SetValue(ctx sdk.Context, accessPath *vm_grpc.VMAccessPath, value []byte) {
}

func (storage VMStorageImpl) GetValue(ctx sdk.Context, accessPath *vm_grpc.VMAccessPath) []byte {
	return nil
}

func (storage VMStorageImpl) HasValue(ctx sdk.Context, accessPath *vm_grpc.VMAccessPath) bool {
	return false
}

func (storage VMStorageImpl) DelValue(ctx sdk.Context, accessPath *vm_grpc.VMAccessPath) {
}

func getMockApp(t *testing.T, numGenAccs int, genState types.GenesisState, genAccs []authexported.Account) testHelper {
	mApp := mock.NewApp()
	types.RegisterCodec(mApp.Cdc)
	keyPricefeed := sdk.NewKVStoreKey(types.StoreKey)

	pk := mApp.ParamsKeeper
	keeper := NewKeeper(keyPricefeed, mApp.Cdc, pk.Subspace(types.DefaultParamspace), NewVMStorage())

	require.NoError(t, mApp.CompleteSetup(keyPricefeed))

	valTokens := sdk.TokensFromConsensusPower(42)
	var (
		addrs    []sdk.AccAddress
		pubKeys  []crypto.PubKey
		privKeys []crypto.PrivKey
	)

	if len(genAccs) == 0 {
		genAccs, addrs, pubKeys, privKeys = mock.CreateGenAccounts(numGenAccs,
			sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, valTokens)))
	}

	mock.SetGenesis(mApp, genAccs)
	return testHelper{mApp, keeper, addrs, pubKeys, privKeys}
}

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
	stdAssetCode string
	stdAssets    types.Assets
	stdNominee   string
}

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

	input.stdAssetCode = "btc_dfi"

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
