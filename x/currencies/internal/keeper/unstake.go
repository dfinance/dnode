package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/dfinance/dnode/x/currencies/internal/types"
)

func (k Keeper) UnstakeCurrency(ctx sdk.Context, staker sdk.AccAddress) error {
	// Get staking denom (usually sxfi).
	stakingDenom := k.stakingKeeper.BondDenom(ctx)

	// Call ForceRemoveDelegator to remove all delegations.
	if err := k.stakingKeeper.ForceRemoveDelegator(ctx, staker); err != nil {
		return sdkErrors.Wrapf(types.ErrForceUnstake, "error during force unstake delegations for %s: %v", staker, err)

	}

	// Check balance and remove sxfi.
	balances := k.bankKeeper.GetCoins(ctx, staker)

	for _, balance := range balances {
		if balance.Denom == stakingDenom {
			// Remove sxfi from balance.
			if err := k.bankKeeper.SetCoins(ctx, staker, balances.Sub(sdk.Coins{balance})); err != nil {
				return sdkErrors.Wrapf(types.ErrNulifyBalance, "error during nullify user sxfi balance for %s: %v", staker, err)

			}

			// Reducing supply.
			// Both keepers (ccs and supply).
			if err := k.ccsKeeper.DecreaseCurrencySupply(ctx, balance); err != nil {
				return sdkErrors.Wrapf(types.ErrInternal, "can't decrease supply with ccsKeeper for %s: %v", staker, err)
			}

			curSupply := k.supplyKeeper.GetSupply(ctx)
			curSupply = curSupply.SetTotal(curSupply.GetTotal().Sub(sdk.Coins{balance}))
			k.supplyKeeper.SetSupply(ctx, curSupply)
		}
	}

	// Ban account.
	k.stakingKeeper.BanAccount(ctx, staker, ctx.BlockHeight())

	return nil
}
