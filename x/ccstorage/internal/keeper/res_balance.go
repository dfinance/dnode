package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/dfinance/dvm-proto/go/vm_grpc"
	"github.com/dfinance/glav"

	"github.com/dfinance/dnode/x/ccstorage/internal/types"
	"github.com/dfinance/dnode/x/common_vm"
)

// SetAccountBalanceResources updates account Balance resources.
func (k Keeper) SetAccountBalanceResources(ctx sdk.Context, acc exported.Account) error {
	k.modulePerms.AutoCheck(types.PermResUpdate)

	filledBalances, emptyBalances, err := k.newBalances(ctx, acc.GetAddress(), acc.GetCoins())
	if err != nil {
		return fmt.Errorf("set balance resources for address %q: %w", acc.GetAddress(), err)
	}

	// update non-empty resources
	for _, balance := range filledBalances {
		bz, err := balance.ResourceBytes()
		if err != nil {
			return fmt.Errorf("set balance resources for address %q (filled): %w", acc.GetAddress().String(), err)
		}
		k.vmKeeper.SetValue(ctx, balance.AccessPath, bz)
	}

	// update empty resources only if resource already exists
	for _, balance := range emptyBalances {
		if k.vmKeeper.HasValue(ctx, balance.AccessPath) {
			bz, err := balance.ResourceBytes()
			if err != nil {
				return fmt.Errorf("set balance resources for address %q (empty): %w", acc.GetAddress().String(), err)
			}
			k.vmKeeper.SetValue(ctx, balance.AccessPath, bz)
		}
	}

	return nil
}

// GetAccountBalanceResources returns account Balance resources.
func (k Keeper) GetAccountBalanceResources(ctx sdk.Context, addr sdk.AccAddress) (types.Balances, error) {
	k.modulePerms.AutoCheck(types.PermRead)

	addrLibra := common_vm.Bech32ToLibra(addr)
	balances := make(types.Balances, 0)

	for _, currency := range k.GetCurrencies(ctx) {
		accessPath := &vm_grpc.VMAccessPath{
			Address: addrLibra,
			Path:    currency.BalancePath(),
		}
		if bz := k.vmKeeper.GetValue(ctx, accessPath); bz != nil {
			balance, err := types.NewBalance(currency.Denom, accessPath, bz)
			if err != nil {
				return nil, fmt.Errorf("get balance resource for address %q: %w", addr.String(), err)
			}
			balances = append(balances, balance)
		}
	}

	return balances, nil
}

// RemoveAccountBalanceResources removes all account balance resource.
func (k Keeper) RemoveAccountBalanceResources(ctx sdk.Context, addr sdk.AccAddress) {
	k.modulePerms.AutoCheck(types.PermResUpdate)

	addrLibra := common_vm.Bech32ToLibra(addr)

	for _, currency := range k.GetCurrencies(ctx) {
		accessPath := &vm_grpc.VMAccessPath{
			Address: addrLibra,
			Path:    currency.BalancePath(),
		}
		k.vmKeeper.DelValue(ctx, accessPath)
	}
}

// newBalance converts sdk.Coin for sdk.AccAddress to Balance.
func (k Keeper) newBalance(addr sdk.AccAddress, coin sdk.Coin) types.Balance {
	return types.Balance{
		Denom: coin.Denom,
		AccessPath: &vm_grpc.VMAccessPath{
			Address: common_vm.Bech32ToLibra(addr),
			Path:    glav.BalanceVector(coin.Denom),
		},
		Resource: types.ResBalance{
			Value: coin.Amount.BigInt(),
		},
	}
}

// newBalances returns two Balance slices depending on account {coins} and all registered currencies;
//   filledBalances: all account {coins}
//   emptyBalances: empty coins which are registered in the module, but aren't found in {coins}
// len(filledBalances + emptyBalances) == len(coins)
func (k Keeper) newBalances(ctx sdk.Context, addr sdk.AccAddress, coins sdk.Coins) (filledBalances, emptyBalances types.Balances, retErr error) {
	// iterate over account coins, mark found, convert to balance and add to write slice
	filledBalances = make(types.Balances, 0, len(coins))
	foundAccDenoms := make(map[string]bool, len(coins))
	for _, coin := range coins {
		balance := k.newBalance(addr, coin)

		filledBalances = append(filledBalances, balance)
		foundAccDenoms[coin.Denom] = true
	}

	// iterate over all registered currencies and if not found above, add to del slice
	emptyBalances = make(types.Balances, 0)
	for _, currency := range k.GetCurrencies(ctx) {
		if !foundAccDenoms[currency.Denom] {
			balance := k.newBalance(addr, sdk.NewCoin(currency.Denom, sdk.ZeroInt()))

			emptyBalances = append(emptyBalances, balance)
		}
	}

	return
}
