// +build unit

package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/secp256k1"
)

// Test set/get account with empty balances.
func TestVMAuthKeeper_SetAccountEmptyBalance(t *testing.T) {
	t.Parallel()

	input := NewTestInput(t)
	keeper, ccsStorage, ctx := input.accountKeeper, input.ccsStorage, input.ctx
	inAcc := input.CreateAccount(t, nil)
	input.accountKeeper.SetAccount(input.ctx, inAcc)

	outAcc := keeper.GetAccount(ctx, inAcc.GetAddress())
	require.Empty(t, outAcc.GetCoins())
	require.EqualValues(t, inAcc, outAcc)

	balances, err := ccsStorage.GetAccountBalanceResources(ctx, inAcc.GetAddress())
	require.NoError(t, err)
	require.Empty(t, balances)
}

// Test set/get new account with different balances.
func TestVMAuthKeeper_SetAccount(t *testing.T) {
	t.Parallel()

	input := NewTestInput(t)
	keeper, ccsStorage, ctx := input.accountKeeper, input.ccsStorage, input.ctx
	acc := input.CreateAccount(t, sdk.NewCoins(sdk.NewCoin("xfi", sdk.OneInt())))
	input.accountKeeper.SetAccount(input.ctx, acc)

	// check one coin exists
	{
		acc = keeper.GetAccount(ctx, acc.GetAddress())
		require.Len(t, acc.GetCoins(), 1)

		balances, err := ccsStorage.GetAccountBalanceResources(ctx, acc.GetAddress())
		require.NoError(t, err)
		require.Len(t, balances, 1)
		require.True(t, acc.GetCoins().IsEqual(balances.Coins()))
	}

	// add one more coin
	{
		coins := acc.GetCoins()
		coins = append(coins, sdk.NewCoin("eth", sdk.NewInt(2)))
		require.NoError(t, acc.SetCoins(coins))
		keeper.SetAccount(ctx, acc)

		acc = keeper.GetAccount(ctx, acc.GetAddress())
		require.Len(t, acc.GetCoins(), 2)
		require.True(t, coins.IsEqual(acc.GetCoins()))

		balances, err := ccsStorage.GetAccountBalanceResources(ctx, acc.GetAddress())
		require.NoError(t, err)
		require.Len(t, balances, 2)
		require.True(t, acc.GetCoins().IsEqual(balances.Coins()))
	}

	// set one coin to zero
	{
		coins := acc.GetCoins()
		coins[0].Amount = sdk.ZeroInt()
		require.NoError(t, acc.SetCoins(coins))
		keeper.SetAccount(ctx, acc)

		// one coin should be returned
		acc = keeper.GetAccount(ctx, acc.GetAddress())
		require.Len(t, acc.GetCoins(), 1)

		// two balances should be returned as resource wasn't removed for zero coin
		balances, err := ccsStorage.GetAccountBalanceResources(ctx, acc.GetAddress())
		require.NoError(t, err)
		require.Len(t, balances, 2)

		// balances.Coins() returns only non-zero coins
		require.True(t, acc.GetCoins().IsEqual(balances.Coins()))
	}

	// change all coins amount and add one more
	// one resource should be added, the rest should be updated
	{
		coins := sdk.NewCoins(
			sdk.NewCoin("xfi", sdk.NewInt(100)),
			sdk.NewCoin("eth", sdk.NewInt(200)),
			sdk.NewCoin("btc", sdk.NewInt(300)),
		)
		require.NoError(t, acc.SetCoins(coins))
		keeper.SetAccount(ctx, acc)

		acc = keeper.GetAccount(ctx, acc.GetAddress())
		require.Len(t, acc.GetCoins(), 3)
		require.True(t, coins.IsEqual(acc.GetCoins()))

		balances, err := ccsStorage.GetAccountBalanceResources(ctx, acc.GetAddress())
		require.NoError(t, err)
		require.Len(t, balances, 3)
		require.True(t, acc.GetCoins().IsEqual(balances.Coins()))
	}

	// update account coins removing some
	{
		coins := acc.GetCoins()
		coins = coins[2:]
		require.NoError(t, acc.SetCoins(coins))
		keeper.SetAccount(ctx, acc)

		acc = keeper.GetAccount(ctx, acc.GetAddress())
		require.Len(t, acc.GetCoins(), len(coins))
		require.True(t, coins.IsEqual(acc.GetCoins()))

		// balances should still have removed coins, but they should be zero
		balances, err := ccsStorage.GetAccountBalanceResources(ctx, acc.GetAddress())
		require.NoError(t, err)
		require.Len(t, balances, 3)

		// balances.Coins() only returns non-zero coins
		require.True(t, coins.IsEqual(balances.Coins()))
	}
}

// Test account resources exists, but account is not.
func TestVMAuthKeeper_GetExistingAccount(t *testing.T) {
	t.Parallel()

	input := NewTestInput(t)
	keeper, ccsStorage, ctx := input.accountKeeper, input.ccsStorage, input.ctx

	// create resources
	coins := sdk.NewCoins(
		sdk.NewCoin("xfi", sdk.NewInt(100)),
		sdk.NewCoin("eth", sdk.NewInt(200)),
		sdk.NewCoin("btc", sdk.NewInt(300)),
	)
	acc := input.CreateAccount(t, coins)
	require.NoError(t, ccsStorage.SetAccountBalanceResources(ctx, acc))

	acc = keeper.GetAccount(ctx, acc.GetAddress())
	require.Len(t, acc.GetCoins(), 3)
	require.True(t, coins.IsEqual(acc.GetCoins()))
}

// Check removing account with balance resources.
func TestVMAuthKeeper_RemoveAccount(t *testing.T) {
	t.Parallel()

	input := NewTestInput(t)
	keeper, ccsStorage, ctx := input.accountKeeper, input.ccsStorage, input.ctx

	coins := sdk.NewCoins(
		sdk.NewCoin("xfi", sdk.NewInt(100)),
		sdk.NewCoin("eth", sdk.NewInt(200)),
		sdk.NewCoin("btc", sdk.NewInt(300)),
	)
	acc := input.CreateAccount(t, coins)
	keeper.SetAccount(ctx, acc)

	// check resources available
	acc = keeper.GetAccount(ctx, acc.GetAddress())
	require.Len(t, acc.GetCoins(), 3)

	balances, err := ccsStorage.GetAccountBalanceResources(ctx, acc.GetAddress())
	require.NoError(t, err)
	require.Len(t, balances, 3)

	// remove account
	addr := acc.GetAddress()
	keeper.RemoveAccount(ctx, acc)

	// check removed
	acc = keeper.GetAccount(ctx, addr)
	require.Nil(t, acc)

	balances, err = ccsStorage.GetAccountBalanceResources(ctx, addr)
	require.NoError(t, err)
	require.Empty(t, balances)
}

// Test get signer acc.
func TestVMAccountKeeper_GetSignerAcc(t *testing.T) {
	t.Parallel()

	input := NewTestInput(t)
	keeper, ctx := input.accountKeeper, input.ctx
	acc := input.CreateAccount(t, nil)
	keeper.SetAccount(ctx, acc)

	outAcc1, err := GetSignerAcc(ctx, keeper, acc.GetAddress())
	require.NoError(t, err)
	require.EqualValues(t, acc, outAcc1)

	outAcc2, err := GetSignerAcc(ctx, keeper, sdk.AccAddress("bmp"))
	require.Error(t, err)
	require.Nil(t, outAcc2)
}

// Test removing balance resource without keeper involvement.
func TestVMAccountKeeper_RemoveBalance(t *testing.T) {
	t.Parallel()

	input := NewTestInput(t)
	keeper, ccsStorage, vmStorage, ctx := input.accountKeeper, input.ccsStorage, input.vmStorage, input.ctx

	coins := sdk.NewCoins(
		sdk.NewCoin("xfi", sdk.NewInt(100)),
		sdk.NewCoin("eth", sdk.NewInt(200)),
	)
	acc := input.CreateAccount(t, coins)
	keeper.SetAccount(ctx, acc)

	balances, err := ccsStorage.GetAccountBalanceResources(ctx, acc.GetAddress())
	require.NoError(t, err)

	// remove eth balance from the VM storage
	for _, b := range balances {
		if b.Denom == "eth" {
			vmStorage.DelValue(ctx, b.AccessPath)
		}
	}

	// balance should be removed from account coins
	acc = keeper.GetAccount(ctx, acc.GetAddress())
	require.Len(t, acc.GetCoins(), 1)
	require.Equal(t, acc.GetCoins()[0].Denom, "xfi")
}

// Test balance resource modification without keeper involvement.
func TestVMAccountKeeper_ModifyBalance(t *testing.T) {
	t.Parallel()

	input := NewTestInput(t)
	keeper, ccsStorage, vmStorage, ctx := input.accountKeeper, input.ccsStorage, input.vmStorage, input.ctx

	coins := sdk.NewCoins(
		sdk.NewCoin("xfi", sdk.NewInt(100)),
		sdk.NewCoin("eth", sdk.NewInt(200)),
	)
	acc := input.CreateAccount(t, coins)
	keeper.SetAccount(ctx, acc)

	balances, err := ccsStorage.GetAccountBalanceResources(ctx, acc.GetAddress())
	require.NoError(t, err)

	// modify eth balance
	for _, b := range balances {
		if b.Denom == "eth" {
			b.Resource.Value = sdk.ZeroInt().BigInt()
			bz, err := b.ResourceBytes()
			require.NoError(t, err)
			vmStorage.SetValue(ctx, b.AccessPath, bz)
		}
	}

	// check keeper "caught" that modification
	acc = keeper.GetAccount(ctx, acc.GetAddress())
	require.Len(t, acc.GetCoins(), 1)
	require.Equal(t, acc.GetCoins()[0].Denom, "xfi")
}

// Check bank - vmauth integration works.
func TestVMAccountKeeper_SendBankKeeper(t *testing.T) {
	t.Parallel()

	input := NewTestInput(t)
	keeper, bank, ccsStorage, ctx := input.accountKeeper, input.bankKeeper, input.ccsStorage, input.ctx

	// coins
	xfiValue := sdk.NewCoin("xfi", sdk.NewInt(100))
	ethValue := sdk.NewCoin("eth", sdk.NewInt(1000))

	xfiHalf := xfiValue
	xfiHalf.Amount = xfiHalf.Amount.QuoRaw(2)

	ethHalf := ethValue
	ethHalf.Amount = ethHalf.Amount.QuoRaw(2)

	// set payer
	payerIntitialCoins := sdk.NewCoins(xfiHalf, ethValue)
	payerAcc := input.CreateAccount(t, payerIntitialCoins)
	keeper.SetAccount(ctx, payerAcc)

	// set payee
	payeeAddr := secp256k1.GenPrivKey().PubKey().Address().Bytes()

	// transfer ethHalf
	{
		transferCoins := sdk.NewCoins(ethHalf)

		// payer loses ethHalf, ethHalf and xfiHalf are left
		payerCurCoins, err := bank.SubtractCoins(ctx, payerAcc.GetAddress(), transferCoins)
		require.NoError(t, err)

		// payee get ethHalf
		payeeCurCoins, err := bank.AddCoins(ctx, payeeAddr, transferCoins)
		require.NoError(t, err)
		require.True(t, payeeCurCoins.IsEqual(transferCoins))

		payerAcc = keeper.GetAccount(ctx, payerAcc.GetAddress())
		require.True(t, payerCurCoins.IsEqual(payerIntitialCoins.Sub(transferCoins)))
	}

	// transfer ethHalf (the rest)
	{
		transferCoins := sdk.NewCoins(ethHalf)

		// payer loses ethHalf (no left), xfiHalf is left
		payerCurCoins, err := bank.SubtractCoins(ctx, payerAcc.GetAddress(), transferCoins)
		require.NoError(t, err)

		// payee get ethHalf, ethValue now
		payeeCurCoins, err := bank.AddCoins(ctx, payeeAddr, transferCoins)
		require.NoError(t, err)
		require.True(t, payeeCurCoins.IsEqual(sdk.NewCoins(ethValue)))

		payerAcc = keeper.GetAccount(ctx, payerAcc.GetAddress())
		require.True(t, payerCurCoins.IsEqual(sdk.NewCoins(xfiHalf)))
	}

	// check eth balance still exists
	{
		balances, err := ccsStorage.GetAccountBalanceResources(ctx, payerAcc.GetAddress())
		require.NoError(t, err)
		require.Len(t, balances, 2)

		for _, balance := range balances {
			switch balance.Denom {
			case "xfi":
				require.Equal(t, xfiHalf.Amount.String(), balance.Resource.Value.String())
			case "eth":
				require.EqualValues(t, balance.Resource.Value.Uint64(), 0)
			default:
				t.Fatalf("unexpected denom found: %s", balance.Denom)
			}
		}
	}
}
