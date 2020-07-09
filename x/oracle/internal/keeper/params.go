package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dfinance/dnode/x/oracle/internal/types"
)

// GetParams gets params from the store.
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	return types.NewParams(k.GetAssetParams(ctx), k.GetNomineeParams(ctx), k.GetPostPriceParams(ctx))
}

// SetParams updates params in the store.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramstore.SetParamSet(ctx, &params)
}

// GetAssetParams get asset params from store.
func (k Keeper) GetAssetParams(ctx sdk.Context) types.Assets {
	var assets types.Assets
	k.paramstore.Get(ctx, types.KeyAssets, &assets)

	return assets
}

// GetNomineeParams get nominee params from store.
func (k Keeper) GetNomineeParams(ctx sdk.Context) []string {
	var nominees []string
	k.paramstore.Get(ctx, types.KeyNominees, &nominees)

	return nominees
}

// GetPostPriceParams get nominee params from store.
func (k Keeper) GetPostPriceParams(ctx sdk.Context) types.PostPriceParams {
	params := types.PostPriceParams{}
	k.paramstore.Get(ctx, types.KeyPostPrice, &params)

	return params
}
