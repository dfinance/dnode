package poa

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/params"
	"wings-blockchain/x/poa/types"
)

type Keeper struct {
	storeKey 	sdk.StoreKey
	cdc 	 	*codec.Codec
	paramStore  params.Subspace
}

func NewKeeper(storeKey sdk.StoreKey, cdc *codec.Codec, paramStore params.Subspace) Keeper {
	return Keeper{
		storeKey:  storeKey,
		cdc:  cdc,
		paramStore: paramStore.WithKeyTable(NewKeyTable()),
	}
}

func (keeper Keeper) AddValidator(ctx sdk.Context, address sdk.AccAddress, ethAddress string) {
	store := ctx.KVStore(keeper.storeKey)

	validator := types.NewValidator(address, ethAddress)
	keeper.increaseValidatorsAmount(ctx)

	store.Set(address, keeper.cdc.MustMarshalBinaryBare(validator))
}

func (keeper Keeper) HasValidator(ctx sdk.Context, address sdk.AccAddress) bool {
	store := ctx.KVStore(keeper.storeKey)

	return store.Has(address)
}

func (keeper Keeper) RemoveValidator(ctx sdk.Context, address sdk.AccAddress) {
	store := ctx.KVStore(keeper.storeKey)

	store.Delete(address)
	keeper.reduceValidatorsAmount(ctx)
}

func (keeper Keeper) ReplaceValidator(ctx sdk.Context, oldAddress sdk.AccAddress, newAddress sdk.AccAddress, ethAddress string)  {
	store := ctx.KVStore(keeper.storeKey)

	store.Delete(oldAddress)

	validator := types.NewValidator(newAddress, ethAddress)
	store.Set(newAddress, keeper.cdc.MustMarshalBinaryBare(validator))
}

func (keeper Keeper) GetValidator(ctx sdk.Context, address sdk.AccAddress) types.Validator {
	store := ctx.KVStore(keeper.storeKey)

	b := store.Get(address)

	var validator types.Validator
	keeper.cdc.MustUnmarshalBinaryBare(b, &validator)

	return validator
}

func (keeper Keeper) GetValidatorAmount(ctx sdk.Context) uint16 {
	store := ctx.KVStore(keeper.storeKey)

	b := store.Get(types.ValidatorsCountKey)
	var amount uint16

	keeper.cdc.MustUnmarshalBinaryBare(b, &amount)

	return amount
}

func (keeper Keeper) increaseValidatorsAmount(ctx sdk.Context) uint16 {
	amount := keeper.GetValidatorAmount(ctx)

	amount += 1
	keeper.setValidatorsAmount(ctx, amount)

	return amount
}

func (keeper Keeper) reduceValidatorsAmount(ctx sdk.Context) uint16 {
	amount := keeper.GetValidatorAmount(ctx)

	amount -= 1

	keeper.setValidatorsAmount(ctx, amount)
	return amount
}

func (keeper Keeper) setValidatorsAmount(ctx sdk.Context, newAmount uint16) {
	store := ctx.KVStore(keeper.storeKey)

	store.Set(types.ValidatorsCountKey, keeper.cdc.MustMarshalBinaryBare(newAmount))
}