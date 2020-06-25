// +build unit

package vmauth

import (
	"math/big"
	"testing"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/secp256k1"
)

// Test set/get account with empty balances.
func TestVMAccountKeeper_SetAccountEmptyBalances(t *testing.T) {
	input := newTestInput(t)

	addr := secp256k1.GenPrivKey().PubKey().Address().Bytes()
	acc := auth.NewBaseAccountWithAddress(addr)
	input.accountKeeper.SetAccount(input.ctx, &acc)

	getter := input.accountKeeper.GetAccount(input.ctx, addr)
	require.EqualValues(t, &acc, getter)

	balances := input.accountKeeper.loadBalances(input.ctx, addr)
	require.Empty(t, balances, "balances not empty")
	require.Empty(t, getter.GetCoins(), "coins not empty")
}

// Test set/get new account with balance.
func TestVMAccountKeeper_SetAccount(t *testing.T) {
	input := newTestInput(t)

	addr := secp256k1.GenPrivKey().PubKey().Address().Bytes()
	acc := auth.NewBaseAccountWithAddress(addr)

	coin := types.NewCoin("dfi", types.NewInt(1))
	coins := types.Coins{coin}
	if err := acc.SetCoins(coins); err != nil {
		t.Fatal(err)
	}

	input.accountKeeper.SetAccount(input.ctx, &acc)
	getter := input.accountKeeper.GetAccount(input.ctx, addr)

	balances := input.accountKeeper.loadBalances(input.ctx, addr)
	require.Len(t, balances, 1, "balances length not equal to 1 (saved one)")
	vmCoins := balancesToCoins(balances)
	require.True(t, getter.GetCoins().IsEqual(vmCoins), "coins are not match after set account")

	// add new resource to vm.
	balances, toDelete := coinsToBalances(getter)
	require.Len(t, toDelete, len(denomPaths)-len(coins)) // contains rest (exclude dfi)
	for _, toDel := range toDelete {
		require.False(t, input.vmStorage.HasValue(input.ctx, toDel.accessPath))
	}

	balances = append(balances, toDelete...)
	input.accountKeeper.saveBalances(input.ctx, balances, nil)
	require.True(t, input.vmStorage.HasValue(input.ctx, toDelete[0].accessPath)) // resource now should exists.

	getter = input.accountKeeper.GetAccount(input.ctx, addr)
	require.Len(t, getter.GetCoins(), 1) // but still doesn't contains eth as it zero value.

	err := AddDenomPath("test1", "00")
	require.NoError(t, err)
	defer RemoveDenomPath("test1")

	balances, toDelete = coinsToBalances(getter)
	require.Len(t, toDelete, len(denomPaths)-len(coins)) // contains 2 - eth and test1

	for i := range toDelete {
		if toDelete[i].denom == "test1" {
			toDelete[i].balance.Value = big.NewInt(100)
			balances = append(balances, toDelete[i])
			break
		}
	}

	input.accountKeeper.saveBalances(input.ctx, balances, nil)
	getter = input.accountKeeper.GetAccount(input.ctx, addr)
	require.Len(t, getter.GetCoins(), 2) // but still doesn't contains eth as it zero value.
	realCoins := balancesToCoins(balances)

	require.True(t, realCoins.IsEqual(getter.GetCoins()))
}

// Check get on account exists only in VM (check how balances works).
func TestVMAccount_GetExistsAccount(t *testing.T) {
	// Save account into.
	input := newTestInput(t)
	addr := secp256k1.GenPrivKey().PubKey().Address().Bytes()
	acc := auth.NewBaseAccountWithAddress(addr)

	// Set coins.
	dfiCoin := types.NewCoin("dfi", types.NewInt(100500))
	ethCoin := types.NewCoin("eth", types.NewInt(100200))
	coins := types.Coins{dfiCoin, ethCoin}

	if err := acc.SetCoins(coins); err != nil {
		t.Fatal(err)
	}

	balances, toDelete := coinsToBalances(&acc)
	require.Len(t, toDelete, len(denomPaths)-len(coins))

	require.Len(t, balances, len(coins), "balances length doesnt match coins")

	input.accountKeeper.saveBalances(input.ctx, balances, toDelete)

	getter := input.accountKeeper.GetAccount(input.ctx, addr)
	realCoins := balancesToCoins(balances)

	require.True(t, realCoins.IsEqual(getter.GetCoins()), "coins doesnt match")
}

// Test save/load balances.
func TestVMAccount_LoadBalances(t *testing.T) {
	input := newTestInput(t)
	addr := secp256k1.GenPrivKey().PubKey().Address().Bytes()

	dfiCoin := types.NewCoin("dfi", types.NewInt(100500))
	ethCoin := types.NewCoin("eth", types.NewInt(100200))
	coins := types.Coins{dfiCoin, ethCoin}

	acc := auth.NewBaseAccountWithAddress(addr)
	if err := acc.SetCoins(coins); err != nil {
		t.Fatal(err)
	}

	balances, toDelete := coinsToBalances(&acc)
	require.Len(t, toDelete, len(denomPaths)-len(coins))

	input.accountKeeper.saveBalances(input.ctx, balances, toDelete)

	loadedBalances := input.accountKeeper.loadBalances(input.ctx, addr)
	realBalances := balancesToCoins(loadedBalances)

	require.True(t, coins.IsEqual(realBalances))
}

// Test remove account.
func TestVMAccountKeeper_RemoveAccount(t *testing.T) {
	input := newTestInput(t)

	addr := secp256k1.GenPrivKey().PubKey().Address().Bytes()
	acc := auth.NewBaseAccountWithAddress(addr)

	input.accountKeeper.SetAccount(input.ctx, &acc)
	input.accountKeeper.RemoveAccount(input.ctx, &acc)

	getter := input.accountKeeper.GetAccount(input.ctx, addr)
	require.Nil(t, getter)

	balances := loadAccessPaths(addr)
	for _, b := range balances {
		bz := input.vmStorage.GetValue(input.ctx, b.accessPath)
		if bz != nil {
			require.Nil(t, bz, "still contains coins balances after removing account")
		}
	}
}

// Test get signer acc.
func TestVMAccountKeeper_GetSignerAcc(t *testing.T) {
	input := newTestInput(t)
	addr := secp256k1.GenPrivKey().PubKey().Address().Bytes()
	acc := auth.NewBaseAccountWithAddress(addr)
	input.accountKeeper.SetAccount(input.ctx, &acc)

	getter, err := GetSignerAcc(input.ctx, input.accountKeeper, addr)
	require.EqualValues(t, &acc, getter)
	require.NoError(t, err)

	getter, err = GetSignerAcc(input.ctx, input.accountKeeper, types.AccAddress("bmp"))
	require.Nil(t, getter)
	require.Error(t, err)
}

// Test balances with toDelete inside SetAccount.
func TestVMAccountKeeper_SetBalancesWithDelete(t *testing.T) {
	input := newTestInput(t)
	addr := secp256k1.GenPrivKey().PubKey().Address().Bytes()
	acc := auth.NewBaseAccountWithAddress(addr)

	err := AddDenomPath("test1", "00")
	require.NoError(t, err)

	err = AddDenomPath("test2", "01")
	require.NoError(t, err)

	defer RemoveDenomPath("test1")
	defer RemoveDenomPath("test2")

	coins := types.Coins{
		types.NewCoin("dfi", types.NewInt(100100)),
		types.NewCoin("eth", types.NewInt(100200)),
		types.NewCoin("btc", types.NewInt(1000)),
		types.NewCoin("usdt", types.NewInt(10)),
		types.NewCoin("test1", types.NewInt(100300)),
		types.NewCoin("test2", types.NewInt(100400)),
	}

	err = acc.SetCoins(coins)
	require.NoError(t, err)

	// just check that there is no toDelete
	_, toDelete := coinsToBalances(&acc)
	require.Empty(t, toDelete)

	input.accountKeeper.SetAccount(input.ctx, &acc)

	coins = types.Coins{
		types.NewCoin("dfi", types.NewInt(100)),
		types.NewCoin("eth", types.NewInt(200)),
	}

	getter := input.accountKeeper.GetAccount(input.ctx, addr)
	require.NotNil(t, getter)
	err = getter.SetCoins(coins)
	require.NoError(t, err)

	input.accountKeeper.SetAccount(input.ctx, getter)
	getter = input.accountKeeper.GetAccount(input.ctx, addr)

	require.True(t, coins.IsEqual(getter.GetCoins()))
}

// Test when remove balance resource from VM.
func TestVMAccountKeeper_RemoveBalance(t *testing.T) {
	input := newTestInput(t)
	addr := secp256k1.GenPrivKey().PubKey().Address().Bytes()
	acc := auth.NewBaseAccountWithAddress(addr)

	coins := types.Coins{
		types.NewCoin("dfi", types.NewInt(100100)),
		types.NewCoin("eth", types.NewInt(100200)),
	}

	err := acc.SetCoins(coins)
	require.NoError(t, err)

	input.accountKeeper.SetAccount(input.ctx, &acc)
	balances := input.accountKeeper.loadBalances(input.ctx, addr)

	// let's say resources for eth removed in the vm.
	for i := range balances {
		if balances[i].denom == "eth" {
			input.vmStorage.DelValue(input.ctx, balances[i].accessPath)

			break
		}
	}

	// get account now and see that account doesn't contains eth anymore.
	getter := input.accountKeeper.GetAccount(input.ctx, addr)
	for _, coin := range getter.GetCoins() {
		require.NotEqual(t, coin.Denom, "eth")
	}
}

// Set ETH balance to zero (like it happened in VM).
func TestVMAccountKeeper_BalanceToZero(t *testing.T) {
	input := newTestInput(t)
	addr := secp256k1.GenPrivKey().PubKey().Address().Bytes()
	acc := auth.NewBaseAccountWithAddress(addr)

	coins := types.Coins{
		types.NewCoin("dfi", types.NewInt(100100)),
		types.NewCoin("eth", types.NewInt(100200)),
	}

	err := acc.SetCoins(coins)
	require.NoError(t, err)

	input.accountKeeper.SetAccount(input.ctx, &acc)
	balances := input.accountKeeper.loadBalances(input.ctx, addr)

	// just remove eth.
	var ethBalance Balance
	for i, _ := range balances {
		if balances[i].denom == "eth" {
			balances[i].balance.Value = types.NewInt(0).BigInt()
			ethBalance = balances[i]
			break
		}
	}

	input.accountKeeper.saveBalances(input.ctx, balances, nil)

	getter := input.accountKeeper.GetAccount(input.ctx, addr)
	for _, balance := range balances {
		t.Logf("%s %s\n", balance.denom, balance.balance.Value.String())
	}
	realCoins := balancesToCoins(balances)

	for _, coin := range realCoins {
		t.Logf("real coins: %s\n", coin.String())
	}

	require.True(t, getter.GetCoins().IsEqual(realCoins))

	// check that eth resource still exists.
	require.True(t, input.vmStorage.HasValue(input.ctx, ethBalance.accessPath))
}

// Bank keeper tests.

// Test send bank keeper.
func TestVMAccountKeeper_SendBankKeeper(t *testing.T) {
	input := newTestInput(t)

	addr := secp256k1.GenPrivKey().PubKey().Address().Bytes()
	recipient := secp256k1.GenPrivKey().PubKey().Address().Bytes()

	acc := auth.NewBaseAccountWithAddress(addr)

	ethValue := types.NewInt(100)
	part := ethValue.QuoRaw(2)

	dfiPart := types.NewCoin("dfi", types.NewInt(100100))
	coins := types.Coins{
		dfiPart,
		types.NewCoin("eth", ethValue),
	}

	err := acc.SetCoins(coins)
	require.NoError(t, err)

	input.accountKeeper.SetAccount(input.ctx, &acc)
	_, toDelete := coinsToBalances(&acc)
	require.Len(t, toDelete, len(denomPaths)-len(coins))

	getter := input.accountKeeper.GetAccount(input.ctx, addr)
	require.True(t, getter.GetCoins().IsEqual(coins))

	toSubstract := types.Coins{types.NewCoin("eth", part)}

	// Withdraw eth from sender.
	{
		senderCoins, err := input.bankKeeper.SubtractCoins(input.ctx, addr, toSubstract)
		require.NoError(t, err)

		// deposit eth to recipient.
		recipientCoins, err := input.bankKeeper.AddCoins(input.ctx, recipient, toSubstract)
		require.NoError(t, err)

		getter = input.accountKeeper.GetAccount(input.ctx, addr)
		require.True(t, senderCoins.IsEqual(coins.Sub(toSubstract)))
		require.True(t, recipientCoins.IsEqual(toSubstract))
	}

	// Withdraw and deposit rest.
	{
		senderCoins, err := input.bankKeeper.SubtractCoins(input.ctx, addr, toSubstract)
		require.NoError(t, err)

		recipientCoins, err := input.bankKeeper.AddCoins(input.ctx, recipient, toSubstract)
		require.NoError(t, err)

		getter = input.accountKeeper.GetAccount(input.ctx, addr)
		require.True(t, senderCoins.IsEqual(types.Coins{dfiPart}))
		require.True(t, recipientCoins.AmountOf("eth").Equal(ethValue))
	}

	// Check that ETH resources still exists.
	{
		balances := input.accountKeeper.loadBalances(input.ctx, addr)
		require.Len(t, balances, 2)

		for _, b := range balances {
			switch b.denom {
			case "dfi":
				require.Equal(t, b.balance.Value.String(), "100100")
				break

			case "eth":
				require.Equal(t, b.balance.Value.String(), "0")
				break

			default:
				t.Fatalf("unknown denom %s", b.denom)
			}
		}
	}
}
