package keeper

import sdk "github.com/cosmos/cosmos-sdk/types"

const (
	stakingDenom = "sxfi"
)

func (k Keeper) UnstakeCurrency(ctx sdk.Context, staker sdk.AccAddress) error {
	// TODO: check user balance and remove sxfi.
	balances := k.bankKeeper.GetCoins(ctx, staker)

	for _, balance := range balances {
		if balance.Denom == stakingDenom {
			// Nullify user sxfi balance.
			err := k.bankKeeper.SetCoins(ctx, staker, balances.Sub(sdk.Coins{balance}))
			if err != nil {
				return err
			}
		}
	}

	// TODO: call ForceRemoveDelegator.
	return nil
}
