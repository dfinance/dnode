package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dfinance/dnode/x/oracle/internal/types"
)

// SetAsset overwrites existing asset for specific assetCode.
func (k Keeper) SetAsset(ctx sdk.Context, nominee string, assetCode string, asset types.Asset) error {
	if !k.IsNominee(ctx, nominee) {
		return fmt.Errorf("%q is not a valid nominee", nominee)
	}

	assets := k.GetAssetParams(ctx)
	updateAssets := assets[:0]
	found := false
	for _, a := range assets {
		if assetCode == a.AssetCode {
			a = asset
			found = true
		}
		updateAssets = append(updateAssets, a)
	}
	if found {
		params := k.GetParams(ctx)
		params.Assets = updateAssets
		k.SetParams(ctx, params)
		return nil
	}

	return fmt.Errorf("asset %q not found", assetCode)
}

// AddAsset adds non-existing asset to the store.
func (k Keeper) AddAsset(ctx sdk.Context, nominee string, assetCode string, asset types.Asset) error {
	// TODO: assetCode input can be obtained from asset.AssetCode input, so might be excessive
	if !k.IsNominee(ctx, nominee) {
		return fmt.Errorf("%q is not a valid nominee", nominee)
	}

	if _, exists := k.GetAsset(ctx, assetCode); exists {
		return fmt.Errorf("asset %q already exists", assetCode)
	}

	assets := k.GetAssetParams(ctx)
	assets = append(assets, asset)

	params := k.GetParams(ctx)
	params.Assets = assets
	k.SetParams(ctx, params)

	return nil
}

// GetAsset returns the asset if it is in the oracle system.
func (k Keeper) GetAsset(ctx sdk.Context, assetCode string) (types.Asset, bool) {
	assets := k.GetAssetParams(ctx)

	for i := range assets {
		if assets[i].AssetCode == assetCode {
			return assets[i], true
		}
	}
	return types.Asset{}, false

}
