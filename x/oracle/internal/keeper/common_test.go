package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authexported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/mock"
	"github.com/dfinance/dvm-proto/go/vm_grpc"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"

	"github.com/dfinance/dnode/x/oracle/internal/keeper"
	"github.com/dfinance/dnode/x/oracle/internal/types"
)

type testHelper struct {
	mApp     *mock.App
	keeper   keeper.Keeper
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

func (storage VMStorageImpl) DelValue(ctx sdk.Context, accessPath *vm_grpc.VMAccessPath) {
}

func getMockApp(t *testing.T, numGenAccs int, genState types.GenesisState, genAccs []authexported.Account) testHelper {
	mApp := mock.NewApp()
	types.RegisterCodec(mApp.Cdc)
	keyPricefeed := sdk.NewKVStoreKey(types.StoreKey)

	pk := mApp.ParamsKeeper
	keeper := keeper.NewKeeper(keyPricefeed, mApp.Cdc, pk.Subspace(types.DefaultParamspace), types.DefaultCodespace, NewVMStorage())

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
