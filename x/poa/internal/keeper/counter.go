package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dfinance/dnode/x/poa/internal/types"
)

// GetValidatorAmount returns current validators amount counter.
func (k Keeper) GetValidatorAmount(ctx sdk.Context) uint16 {
	k.modulePerms.AutoCheck(types.PermRead)

	store := ctx.KVStore(k.storeKey)

	if !store.Has(types.ValidatorsCountKey) {
		return 0
	}

	var amount uint16
	bz := store.Get(types.ValidatorsCountKey)
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &amount)

	return amount
}

// GetEnoughConfirmations returns minimal number of votes (confirmations) to perform a multi signature action.
func (k Keeper) GetEnoughConfirmations(ctx sdk.Context) uint16 {
	k.modulePerms.AutoCheck(types.PermRead)

	return k.GetValidatorAmount(ctx)/2 + 1
}

// increaseValidatorsAmount increases current validators amount counter by 1.
func (k Keeper) increaseValidatorsAmount(ctx sdk.Context) uint16 {
	amount := k.GetValidatorAmount(ctx)

	amount += 1
	k.setValidatorsAmount(ctx, amount)

	return amount
}

// decreaseValidatorsAmount decreases current validators amount counter by 1.
func (k Keeper) decreaseValidatorsAmount(ctx sdk.Context) uint16 {
	amount := k.GetValidatorAmount(ctx)

	amount -= 1
	k.setValidatorsAmount(ctx, amount)

	return amount
}

// setValidatorsAmount sets validators counter to the storage.
func (k Keeper) setValidatorsAmount(ctx sdk.Context, amount uint16) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.ValidatorsCountKey, k.cdc.MustMarshalBinaryLengthPrefixed(amount))
}
