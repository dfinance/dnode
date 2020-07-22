package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dfinance/dnode/x/poa/internal/types"
)

// GetMaxValidators returns maxValidators param.
func (k Keeper) GetMaxValidators(ctx sdk.Context) (res uint16) {
	k.modulePerms.AutoCheck(types.PermRead)

	k.paramStore.Get(ctx, types.ParamStoreKeyMaxValidators, &res)
	return
}

// Get minimum validators amount.
func (k Keeper) GetMinValidators(ctx sdk.Context) (res uint16) {
	k.modulePerms.AutoCheck(types.PermRead)

	k.paramStore.Get(ctx, types.ParamStoreKeyMinValidators, &res)
	return
}

// GetParams returns keeper params.
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	k.modulePerms.AutoCheck(types.PermRead)

	min := k.GetMinValidators(ctx)
	max := k.GetMaxValidators(ctx)

	return types.NewParams(max, min)
}

// setParams sets keeper params.
func (k Keeper) setParams(ctx sdk.Context, params types.Params) {
	k.paramStore.SetParamSet(ctx, &params)
}
