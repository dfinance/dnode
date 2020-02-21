// PoA keeper main functional.
package poa

import (
	"github.com/WingsDao/wings-blockchain/x/poa/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
)

// PoA keeper implementation.
type Keeper struct {
	storeKey   sdk.StoreKey
	cdc        *codec.Codec
	paramStore params.Subspace
}

// Creating new keeper with parameters store.
func NewKeeper(storeKey sdk.StoreKey, cdc *codec.Codec, paramStore params.Subspace) Keeper {
	return Keeper{
		storeKey:   storeKey,
		cdc:        cdc,
		paramStore: paramStore.WithKeyTable(NewKeyTable()),
	}
}

// Add new validator to list of PoA validators.
func (poaKeeper Keeper) AddValidator(ctx sdk.Context, address sdk.AccAddress, ethAddress string) {
	store := ctx.KVStore(poaKeeper.storeKey)

	validator := types.NewValidator(address, ethAddress)
	poaKeeper.increaseValidatorsAmount(ctx)

	store.Set(address, poaKeeper.cdc.MustMarshalBinaryBare(validator))
	poaKeeper.addValidatorToList(ctx, validator)
}

// Get validators from validators list.
func (poaKeeper Keeper) GetValidators(ctx sdk.Context) types.Validators {
	return poaKeeper.getValidatorsList(ctx)
}

// Check if validator exists in list or not.
func (poaKeeper Keeper) HasValidator(ctx sdk.Context, address sdk.AccAddress) bool {
	store := ctx.KVStore(poaKeeper.storeKey)

	return store.Has(address)
}

// Remove validator that exists in validators list.
func (poaKeeper Keeper) RemoveValidator(ctx sdk.Context, address sdk.AccAddress) {
	store := ctx.KVStore(poaKeeper.storeKey)

	store.Delete(address)
	poaKeeper.reduceValidatorsAmount(ctx)
	poaKeeper.removeValidatorFromList(ctx, address)
}

// Replace validator with another one.
func (poaKeeper Keeper) ReplaceValidator(ctx sdk.Context, oldAddress sdk.AccAddress, newAddress sdk.AccAddress, ethAddress string) {
	store := ctx.KVStore(poaKeeper.storeKey)

	store.Delete(oldAddress)

	validator := types.NewValidator(newAddress, ethAddress)
	store.Set(newAddress, poaKeeper.cdc.MustMarshalBinaryBare(validator))
}

// Getting validator from storage.
func (poaKeeper Keeper) GetValidator(ctx sdk.Context, address sdk.AccAddress) types.Validator {
	store := ctx.KVStore(poaKeeper.storeKey)

	b := store.Get(address)

	var validator types.Validator
	poaKeeper.cdc.MustUnmarshalBinaryBare(b, &validator)

	return validator
}

// Get total amount of validators.
func (poaKeeper Keeper) GetValidatorAmount(ctx sdk.Context) uint16 {
	store := ctx.KVStore(poaKeeper.storeKey)

	if !store.Has(types.ValidatorsCountKey) {
		return 0
	}

	b := store.Get(types.ValidatorsCountKey)
	var amount uint16

	poaKeeper.cdc.MustUnmarshalBinaryLengthPrefixed(b, &amount)

	return amount
}

// Get amount of confirmations to do action.
func (poaKeeper Keeper) GetEnoughConfirmations(ctx sdk.Context) uint16 {
	return poaKeeper.GetValidatorAmount(ctx)/2 + 1
}

// Add validator to validators list.
func (poaKeeper Keeper) addValidatorToList(ctx sdk.Context, validator types.Validator) {
	validators := poaKeeper.getValidatorsList(ctx)
	validators = append(validators, validator)
	poaKeeper.storeValidatorsList(ctx, validators)
}

// Remove validator from validator list by address.
func (poaKeeper Keeper) removeValidatorFromList(ctx sdk.Context, address sdk.AccAddress) {
	validators := poaKeeper.getValidatorsList(ctx)

	index := -1

	for i, validator := range validators {
		if validator.Address.Equals(address) {
			index = i
			break
		}
	}

	if index >= 0 {
		if len(validators) > 1 {
			validators = append(validators[:index], validators[index+1:]...)
			poaKeeper.storeValidatorsList(ctx, validators)
		} else {
			store := ctx.KVStore(poaKeeper.storeKey)
			store.Delete(types.ValidatorsListKey)
		}
	}
}

// Get validators list.
func (poaKeeper Keeper) getValidatorsList(ctx sdk.Context) types.Validators {
	store := ctx.KVStore(poaKeeper.storeKey)

	if !store.Has(types.ValidatorsListKey) {
		return types.Validators{}
	}

	var validators types.Validators
	bs := store.Get(types.ValidatorsListKey)

	err := poaKeeper.cdc.UnmarshalBinaryBare(bs, &validators)

	if err != nil {
		panic(err)
	}

	return validators
}

// Store validators list.
func (poaKeeper Keeper) storeValidatorsList(ctx sdk.Context, validators types.Validators) {
	store := ctx.KVStore(poaKeeper.storeKey)
	store.Set(types.ValidatorsListKey, poaKeeper.cdc.MustMarshalBinaryBare(validators))
}

// Increase validators amount by 1.
func (poaKeeper Keeper) increaseValidatorsAmount(ctx sdk.Context) uint16 {
	amount := poaKeeper.GetValidatorAmount(ctx)

	amount += 1
	poaKeeper.setValidatorsAmount(ctx, amount)

	return amount
}

// Reduce validators amount by 1.
func (poaKeeper Keeper) reduceValidatorsAmount(ctx sdk.Context) uint16 {
	amount := poaKeeper.GetValidatorAmount(ctx)

	amount -= 1

	poaKeeper.setValidatorsAmount(ctx, amount)
	return amount
}

// Set new validators amount.
func (poaKeeper Keeper) setValidatorsAmount(ctx sdk.Context, newAmount uint16) {
	store := ctx.KVStore(poaKeeper.storeKey)

	store.Set(types.ValidatorsCountKey, poaKeeper.cdc.MustMarshalBinaryLengthPrefixed(newAmount))
}
