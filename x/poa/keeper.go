package poa

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/params"
	"wings-blockchain/x/poa/types"
)

// PoA keeper implementation
type Keeper struct {
	storeKey 	sdk.StoreKey
	cdc 	 	*codec.Codec
	paramStore  params.Subspace
}

// Creating new keeper with parameters store
func NewKeeper(storeKey sdk.StoreKey, cdc *codec.Codec, paramStore params.Subspace) Keeper {
	return Keeper{
		storeKey:  storeKey,
		cdc:  cdc,
		paramStore: paramStore.WithKeyTable(NewKeyTable()),
	}
}

// Add new validator to list of PoA validators
func (keeper Keeper) AddValidator(ctx sdk.Context, address sdk.AccAddress, ethAddress string) {
	store := ctx.KVStore(keeper.storeKey)

	validator := types.NewValidator(address, ethAddress)
	keeper.increaseValidatorsAmount(ctx)

	store.Set(address, keeper.cdc.MustMarshalBinaryBare(validator))
	keeper.addValidatorToList(ctx, validator)
}

func (keeper Keeper) GetValidators(ctx sdk.Context) types.Validators  {
	return keeper.getValidatorsList(ctx)
}

// Check if validator exists in list or not
func (keeper Keeper) HasValidator(ctx sdk.Context, address sdk.AccAddress) bool {
	store := ctx.KVStore(keeper.storeKey)

	return store.Has(address)
}

// Remove validator that exists in validators list
func (keeper Keeper) RemoveValidator(ctx sdk.Context, address sdk.AccAddress) {
	store := ctx.KVStore(keeper.storeKey)

	store.Delete(address)
	keeper.reduceValidatorsAmount(ctx)
	keeper.removeValidatorFromList(ctx, address)
}

// Replace validator with another one
func (keeper Keeper) ReplaceValidator(ctx sdk.Context, oldAddress sdk.AccAddress, newAddress sdk.AccAddress, ethAddress string)  {
	store := ctx.KVStore(keeper.storeKey)

	store.Delete(oldAddress)

	validator := types.NewValidator(newAddress, ethAddress)
	store.Set(newAddress, keeper.cdc.MustMarshalBinaryBare(validator))
}

// Getting validator from storage
func (keeper Keeper) GetValidator(ctx sdk.Context, address sdk.AccAddress) types.Validator {
	store := ctx.KVStore(keeper.storeKey)

	b := store.Get(address)

	var validator types.Validator
	keeper.cdc.MustUnmarshalBinaryBare(b, &validator)

	return validator
}

// Get total amount of validators
func (keeper Keeper) GetValidatorAmount(ctx sdk.Context) uint16 {
	store := ctx.KVStore(keeper.storeKey)

	if !store.Has(types.ValidatorsCountKey) {
		return 0
	}

	b := store.Get(types.ValidatorsCountKey)
	var amount uint16

	keeper.cdc.MustUnmarshalBinaryLengthPrefixed(b, &amount)

	return amount
}

// Get amount of confirmations to do action
func (keeper Keeper) GetEnoughConfirmations(ctx sdk.Context) uint16 {
	return keeper.GetValidatorAmount(ctx) / 2 + 1
}

// Get codec
func (keeper Keeper) GetCDC() *codec.Codec {
	return keeper.cdc
}

// Add validator to validators list
func (keeper Keeper) addValidatorToList(ctx sdk.Context, validator types.Validator) {
	validators := keeper.getValidatorsList(ctx)
	validators = append(validators, validator)
	keeper.storeValidatorsList(ctx, validators)
}

func (keeper Keeper) removeValidatorFromList(ctx sdk.Context, address sdk.AccAddress) {
	validators := keeper.getValidatorsList(ctx)

	index := -1

	for i, validator := range validators {
		if validator.Address.Equals(address) {
			index = i
			break
		}
	}

	if index >= 0 {
		if len(validators) > 0 {
			validators = append(validators[:index], validators[index+1:]...)
			keeper.storeValidatorsList(ctx, validators)
		}  else {
			store := ctx.KVStore(keeper.storeKey)
			store.Delete(types.ValidatorsListKey)
		}
	}
}

func (keeper Keeper) getValidatorsList(ctx sdk.Context) types.Validators {
	store := ctx.KVStore(keeper.storeKey)

	if !store.Has(types.ValidatorsListKey) {
		return types.Validators{}
	}

	var validators types.Validators
	bs := store.Get(types.ValidatorsListKey)

	err := keeper.cdc.UnmarshalBinaryBare(bs, &validators)

	if err != nil {
		panic(err)
	}

	return validators
}

func (keeper Keeper) storeValidatorsList(ctx sdk.Context, validators types.Validators) {
	store := ctx.KVStore(keeper.storeKey)
	store.Set(types.ValidatorsListKey, keeper.cdc.MustMarshalBinaryBare(validators))
}

// Increase validators amount by 1
func (keeper Keeper) increaseValidatorsAmount(ctx sdk.Context) uint16 {
	amount := keeper.GetValidatorAmount(ctx)

	amount += 1
	keeper.setValidatorsAmount(ctx, amount)

	return amount
}

// Reduce validators amount by 1
func (keeper Keeper) reduceValidatorsAmount(ctx sdk.Context) uint16 {
	amount := keeper.GetValidatorAmount(ctx)

	amount -= 1

	keeper.setValidatorsAmount(ctx, amount)
	return amount
}

// Set new validators amount
func (keeper Keeper) setValidatorsAmount(ctx sdk.Context, newAmount uint16) {
	store := ctx.KVStore(keeper.storeKey)

	store.Set(types.ValidatorsCountKey, keeper.cdc.MustMarshalBinaryLengthPrefixed(newAmount))
}