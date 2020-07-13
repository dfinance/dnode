package keeper

import (
	"fmt"
	dnTypes "github.com/dfinance/dnode/helpers/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dfinance/dnode/x/oracle/internal/types"
)

// GetAsset returns the asset if it is in the oracle system.
func (k Keeper) GetAsset(ctx sdk.Context, assetCode dnTypes.AssetCode) (types.Asset, bool) {
	assets := k.GetAssetParams(ctx)

	for i := range assets {
		if assets[i].AssetCode == assetCode {
			return assets[i], true
		}
	}
	return types.Asset{}, false

}

// SetAsset overwrites existing asset for specific assetCode.
func (k Keeper) SetAsset(ctx sdk.Context, nominee string, asset types.Asset) error {
	if !k.IsNominee(ctx, nominee) {
		return fmt.Errorf("%q is not a valid nominee", nominee)
	}

	assets := k.GetAssetParams(ctx)
	updateAssets := assets[:0]
	found := false
	for _, a := range assets {
		if asset.AssetCode == a.AssetCode {
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

	return fmt.Errorf("asset %q not found", asset.AssetCode)
}

// AddAsset adds non-existing asset to the store.
func (k Keeper) AddAsset(ctx sdk.Context, nominee string, asset types.Asset) error {
	if !k.IsNominee(ctx, nominee) {
		return fmt.Errorf("%q is not a valid nominee", nominee)
	}

	if _, exists := k.GetAsset(ctx, asset.AssetCode); exists {
		return fmt.Errorf("asset %q already exists", asset.AssetCode)
	}

	assets := k.GetAssetParams(ctx)
	assets = append(assets, asset)

	params := k.GetParams(ctx)
	params.Assets = assets
	k.SetParams(ctx, params)

	return nil
}
