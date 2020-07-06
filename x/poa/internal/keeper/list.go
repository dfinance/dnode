package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dfinance/dnode/x/poa/internal/types"
)

// GetValidators returns validators list.
func (k Keeper) GetValidators(ctx sdk.Context) types.Validators {
	store := ctx.KVStore(k.storeKey)

	if !store.Has(types.ValidatorsListKey) {
		return types.Validators{}
	}

	var validators types.Validators
	bz := store.Get(types.ValidatorsListKey)
	k.cdc.MustUnmarshalBinaryBare(bz, &validators)

	return validators
}

// storeValidatorsList stores validators list to the storage.
func (k Keeper) storeValidatorsList(ctx sdk.Context, validators types.Validators) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryBare(validators)
	store.Set(types.ValidatorsListKey, bz)
}

// addValidatorToList adds validator to the list and rewrites list in the storage.
func (k Keeper) addValidatorToList(ctx sdk.Context, validator types.Validator) {
	validators := k.GetValidators(ctx)
	validators = append(validators, validator)
	k.storeValidatorsList(ctx, validators)
}

// removeValidatorFromList removes validator from the list and rewrites list in the storage.
func (k Keeper) removeValidatorFromList(ctx sdk.Context, address sdk.AccAddress) {
	validators := k.GetValidators(ctx)

	idx := -1
	for i, validator := range validators {
		if validator.Address.Equals(address) {
			idx = i
			break
		}
	}

	if idx >= 0 {
		if len(validators) > 1 {
			validators = append(validators[:idx], validators[idx+1:]...)
			k.storeValidatorsList(ctx, validators)
		} else {
			store := ctx.KVStore(k.storeKey)
			store.Delete(types.ValidatorsListKey)
		}
	}
}
