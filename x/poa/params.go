package poa

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	"wings-blockchain/x/poa/types"
)

// New Paramstore for PoA module
func NewKeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&types.Params{})
}

// Get max validators amount
func (poaKeeper Keeper) GetMaxValidators(ctx sdk.Context) (res uint16) {
	poaKeeper.paramStore.Get(ctx, types.KeyMaxValidators, &res)
	return
}

// Get minimum validators amount
func (poaKeeper Keeper) GetMinValidators(ctx sdk.Context) (res uint16) {
	poaKeeper.paramStore.Get(ctx, types.KeyMinValidators, &res)
	return
}

// Get params
func (poaKeeper Keeper) GetParams(ctx sdk.Context) types.Params {
	min := poaKeeper.GetMinValidators(ctx)
	max := poaKeeper.GetMaxValidators(ctx)

	return types.NewParams(min, max)
}

// set the params
func (poaKeeper Keeper) SetParams(ctx sdk.Context, params types.Params) {
	poaKeeper.paramStore.SetParamSet(ctx, &params)
}
