package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/dfinance/dnode/x/poa/internal/types"
)

// AddValidator add a new PoA validator to the list.
func (k Keeper) AddValidator(ctx sdk.Context, address sdk.AccAddress, ethAddress string) error {
	k.modulePerms.AutoCheck(types.PermWriter)

	return k.addValidator(ctx, address, ethAddress, true)
}

// RemoveValidator removes existing PoA validator from the list.
func (k Keeper) RemoveValidator(ctx sdk.Context, address sdk.AccAddress) error {
	k.modulePerms.AutoCheck(types.PermWriter)

	return k.removeValidator(ctx, address, true)
}

// ReplaceValidator removes "old" PoA validator and adds "new" validator to the list.
func (k Keeper) ReplaceValidator(ctx sdk.Context, oldAddress sdk.AccAddress, newAddress sdk.AccAddress, ethAddress string) error {
	k.modulePerms.AutoCheck(types.PermWriter)

	if err := k.removeValidator(ctx, oldAddress, false); err != nil {
		return sdkErrors.Wrap(err, "removing old validator")
	}

	if err := k.addValidator(ctx, newAddress, ethAddress, false); err != nil {
		return sdkErrors.Wrap(err, "adding new validator")
	}

	return nil
}

// HasValidator checks if validator exists.
func (k Keeper) HasValidator(ctx sdk.Context, address sdk.AccAddress) bool {
	k.modulePerms.AutoCheck(types.PermReader)

	store := ctx.KVStore(k.storeKey)

	return store.Has(address)
}

// GetValidator returns validator.
func (k Keeper) GetValidator(ctx sdk.Context, address sdk.AccAddress) (types.Validator, error) {
	k.modulePerms.AutoCheck(types.PermReader)

	if !k.HasValidator(ctx, address) {
		return types.Validator{}, sdkErrors.Wrap(types.ErrValidatorNotExists, address.String())
	}

	return k.getValidator(ctx, address), nil
}

// addValidator add a new PoA validator to the storage and updates stored list.
func (k Keeper) addValidator(ctx sdk.Context, address sdk.AccAddress, ethAddress string, checkLimits bool) error {
	validator := types.NewValidator(address, ethAddress)
	if err := validator.Validate(); err != nil {
		return err
	}

	if k.HasValidator(ctx, address) {
		return sdkErrors.Wrap(types.ErrValidatorExists, address.String())
	}

	if checkLimits {
		maxValidators := k.GetMaxValidators(ctx)
		curValidators := k.GetValidatorAmount(ctx)
		if curValidators+1 > maxValidators {
			return sdkErrors.Wrapf(types.ErrMaxValidatorsReached, "%d", maxValidators)
		}
	}

	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryBare(validator)
	store.Set(validator.Address, bz)

	k.addValidatorToList(ctx, validator)
	k.increaseValidatorsAmount(ctx)

	ctx.EventManager().EmitEvent(types.NewValidatorAddedEvent(validator))

	return nil
}

// removeValidator removes existing PoA validator from the storage and updates stored list.
func (k Keeper) removeValidator(ctx sdk.Context, address sdk.AccAddress, checkLimits bool) error {
	if !k.HasValidator(ctx, address) {
		return sdkErrors.Wrap(types.ErrValidatorNotExists, address.String())
	}

	validator := k.getValidator(ctx, address)

	if checkLimits {
		minValidators := k.GetMinValidators(ctx)
		curValidators := k.GetValidatorAmount(ctx)
		if curValidators-1 < minValidators {
			return sdkErrors.Wrapf(types.ErrMinValidatorsReached, "%d", minValidators)
		}
	}

	store := ctx.KVStore(k.storeKey)
	store.Delete(address)

	k.removeValidatorFromList(ctx, address)
	k.decreaseValidatorsAmount(ctx)

	ctx.EventManager().EmitEvent(types.NewValidatorRemovedEvent(validator))

	return nil
}

// getValidator returns validator from the storage.
func (k Keeper) getValidator(ctx sdk.Context, address sdk.AccAddress) types.Validator {
	store := ctx.KVStore(k.storeKey)

	var validator types.Validator
	bz := store.Get(address)
	k.cdc.MustUnmarshalBinaryBare(bz, &validator)

	return validator
}
