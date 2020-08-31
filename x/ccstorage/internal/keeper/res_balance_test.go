// +build unit

package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/dfinance/dvm-proto/go/vm_grpc"
	"github.com/dfinance/glav"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	"github.com/dfinance/dnode/x/ccstorage/internal/types"
	"github.com/dfinance/dnode/x/common_vm"
)

type balanceInput struct {
	Denom      string
	AccessPath *vm_grpc.VMAccessPath
	Amount     sdk.Int
}

type balanceInputs []balanceInput

func (b balanceInput) CheckBalance(t *testing.T, comment string, found bool, balance types.Balance) {
	require.Falsef(t, found, "%s (%s): duplicated", b.Denom, comment)

	require.Equalf(t, b.Amount.String(), balance.Resource.Value.String(), "%s (%s): amount", b.Denom, comment)
	require.EqualValuesf(t, b.AccessPath.Address, balance.AccessPath.Address, "%s (%s): pathAddr", b.Denom, comment)
	require.EqualValuesf(t, b.AccessPath.Path, balance.AccessPath.Path, "%s (%s): pathPath", b.Denom, comment)
}

func (b balanceInput) CheckResource(t *testing.T, ctx sdk.Context, storage common_vm.VMStorage, exists bool) {
	require.Equalf(t, storage.HasValue(ctx, b.AccessPath), exists, "%s: resource presence", b.Denom)

	if exists {
		bz := storage.GetValue(ctx, b.AccessPath)
		res, err := types.NewResBalance(bz)
		require.NoErrorf(t, err, b.Denom, "%s: resource unmarshal", b.Denom)

		require.Equalf(t, b.Amount.String(), res.Value.String(), "%s: resource value", b.Denom)
	}
}

func (l balanceInputs) Coins() sdk.Coins {
	coins := sdk.Coins{}
	for _, b := range l {
		if b.Amount.GT(sdk.ZeroInt()) {
			coins = append(coins, sdk.Coin{Denom: b.Denom, Amount: b.Amount})
		}
	}

	return coins
}

func (l balanceInputs) AccountWithCoins(t *testing.T, acc exported.Account) exported.Account {
	require.NoError(t, acc.SetCoins(l.Coins()))

	return acc
}

func (l balanceInputs) CheckNewBalances(t *testing.T, filledBalances, emptyBalances types.Balances, err error) {
	require.NoError(t, err, "newBalances error")
	require.Equal(t, len(l), len(filledBalances)+len(emptyBalances), "length mismatch")

	for _, input := range l {
		found := false
		for _, filledBalance := range filledBalances {
			if filledBalance.Denom == input.Denom {
				input.CheckBalance(t, "filled", found, filledBalance)
				found = true
			}
		}
		for _, emptyBalance := range emptyBalances {
			if emptyBalance.Denom == input.Denom {
				input.CheckBalance(t, "empty", found, emptyBalance)
				found = true
			}
		}
		require.Truef(t, found, "%s: not found", input.Denom)
	}
}

func (l balanceInputs) CheckResources(t *testing.T, ctx sdk.Context, storage common_vm.VMStorage, balances types.Balances, err error) {
	require.NoError(t, err, "getResources error")

	for _, input := range l {
		if input.Amount.IsZero() {
			input.CheckResource(t, ctx, storage, false)
			continue
		}

		found := false
		for _, balance := range balances {
			if balance.Denom == input.Denom {
				input.CheckBalance(t, "_", found, balance)
				input.CheckResource(t, ctx, storage, true)
				found = true
			}
		}
		require.Truef(t, found, "%s: not found", input.Denom)
	}
}

func newBalanceInputs(t *testing.T, addr sdk.AccAddress) balanceInputs {
	inputs := balanceInputs{}
	for _, params := range types.DefaultGenesisState().CurrenciesParams {
		currency := types.NewCurrency(params, sdk.ZeroInt())

		inputs = append(inputs, balanceInput{
			Denom: params.Denom,
			AccessPath: &vm_grpc.VMAccessPath{
				Address: common_vm.Bech32ToLibra(addr),
				Path:    currency.BalancePath(),
			},
			Amount: sdk.ZeroInt(),
		})
	}
	require.True(t, len(inputs) >= 2, "genesis doesn't contain at least 2 currencies, test are irrelevant")

	return inputs
}

// Test checks keeper newBalance method.
func TestCCSKeeper_newBalance(t *testing.T) {
	t.Parallel()

	input := NewTestInput(t)
	keeper := input.keeper

	addr := sdk.AccAddress("addr1")
	coin := sdk.Coin{Amount: sdk.NewIntFromUint64(100)}

	// pick an existing currency
	for _, params := range types.DefaultGenesisState().CurrenciesParams {
		coin.Denom = params.Denom
		break
	}

	// ok
	{
		balance := keeper.newBalance(addr, coin)
		path := glav.BalanceVector(coin.Denom)

		require.Equal(t, coin.Denom, balance.Denom)
		require.Equal(t, coin.Amount.String(), balance.Resource.Value.String())
		require.EqualValues(t, common_vm.Bech32ToLibra(addr), balance.AccessPath.Address)
		require.EqualValues(t, path, balance.AccessPath.Path)
	}
}

// Test checks keeper newBalances method.
func TestCCSKeeper_newBalances(t *testing.T) {
	t.Parallel()

	input := NewTestInput(t)
	ctx, keeper := input.ctx, input.keeper
	addr := sdk.AccAddress("addr1")

	inputs := newBalanceInputs(t, addr)

	// all empty
	{
		filledBalances, emptyBalances, err := keeper.newBalances(ctx, addr, inputs.Coins())
		inputs.CheckNewBalances(t, filledBalances, emptyBalances, err)
		require.Len(t, filledBalances, 0)
		require.Len(t, emptyBalances, len(inputs))
	}

	// half filled, half empty
	{
		for i := 0; i < len(inputs)/2; i++ {
			inputs[i].Amount = sdk.NewIntFromUint64(uint64(i) + 100)
		}

		filledBalances, emptyBalances, err := keeper.newBalances(ctx, addr, inputs.Coins())
		inputs.CheckNewBalances(t, filledBalances, emptyBalances, err)
		require.Len(t, filledBalances, len(inputs)/2)
		require.Len(t, emptyBalances, len(inputs)-len(inputs)/2)
	}

	// all filled
	{
		for i := 0; i < len(inputs); i++ {
			inputs[i].Amount = sdk.NewIntFromUint64(uint64(i) + 100)
		}

		filledBalances, emptyBalances, err := keeper.newBalances(ctx, addr, inputs.Coins())
		inputs.CheckNewBalances(t, filledBalances, emptyBalances, err)
		require.Len(t, filledBalances, len(inputs))
		require.Len(t, emptyBalances, 0)
	}

	// all filled, change amounts
	{
		for i := 0; i < len(inputs); i++ {
			inputs[i].Amount = sdk.NewIntFromUint64(uint64(i) + 200)

			filledBalances, emptyBalances, err := keeper.newBalances(ctx, addr, inputs.Coins())
			inputs.CheckNewBalances(t, filledBalances, emptyBalances, err)
			require.Len(t, filledBalances, len(inputs))
			require.Len(t, emptyBalances, 0)
		}
	}
}

// Test checks SetAccountBalanceResources / GetAccountBalanceResources / RemoveAccountBalanceResources keeper methods.
func TestCCSKeeper_AccountBalanceResources(t *testing.T) {
	t.Parallel()

	input := NewTestInput(t)
	ctx, keeper, storage := input.ctx, input.keeper, input.vmStorage
	addr := secp256k1.GenPrivKey().PubKey().Address().Bytes()
	acc := auth.NewBaseAccountWithAddress(addr)

	inputs := newBalanceInputs(t, addr)

	// no balances
	{
		balances, err := keeper.GetAccountBalanceResources(ctx, addr)
		require.NoError(t, err)
		require.Len(t, balances, 0)
	}

	// add balance (total: 1)
	{
		inputs[0].Amount = sdk.NewIntFromUint64(100)

		err := keeper.SetAccountBalanceResources(ctx, inputs.AccountWithCoins(t, &acc))
		require.NoError(t, err)

		balances, err := keeper.GetAccountBalanceResources(ctx, acc.GetAddress())
		inputs.CheckResources(t, ctx, storage, balances, err)
		require.Len(t, balances, 1)
	}

	// add balance (total: 2)
	{
		inputs[1].Amount = sdk.NewIntFromUint64(200)

		err := keeper.SetAccountBalanceResources(ctx, inputs.AccountWithCoins(t, &acc))
		require.NoError(t, err)

		balances, err := keeper.GetAccountBalanceResources(ctx, acc.GetAddress())
		inputs.CheckResources(t, ctx, storage, balances, err)
		require.Len(t, balances, 2)
	}

	// all balances (with updates)
	{
		for i := 0; i < len(inputs); i++ {
			inputs[i].Amount = sdk.NewIntFromUint64(uint64(i) + 1000)
		}

		err := keeper.SetAccountBalanceResources(ctx, inputs.AccountWithCoins(t, &acc))
		require.NoError(t, err)

		balances, err := keeper.GetAccountBalanceResources(ctx, acc.GetAddress())
		inputs.CheckResources(t, ctx, storage, balances, err)
		require.Len(t, balances, len(inputs))
	}

	// remove all resources
	{
		for i := 0; i < len(inputs); i++ {
			inputs[i].Amount = sdk.ZeroInt()
		}

		keeper.RemoveAccountBalanceResources(ctx, addr)

		balances, err := keeper.GetAccountBalanceResources(ctx, acc.GetAddress())
		inputs.CheckResources(t, ctx, storage, balances, err)
		require.Len(t, balances, 0)
	}
}
