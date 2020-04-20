package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dfinance/dnode/x/pricefeed/internal/types"
)

// GetParams gets params from the store
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	return types.NewParams(k.GetAssetParams(ctx), k.GetNomineeParams(ctx), k.GetPostPriceParams(ctx))
}

// SetParams updates params in the store
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramstore.SetParamSet(ctx, &params)
}

// GetAssetParams get asset params from store
func (k Keeper) GetAssetParams(ctx sdk.Context) types.Assets {
	var assets types.Assets
	k.paramstore.Get(ctx, types.KeyAssets, &assets)

	return assets
}

// GetNomineeParams get nominee params from store
func (k Keeper) GetNomineeParams(ctx sdk.Context) []string {
	var nominees []string
	k.paramstore.Get(ctx, types.KeyNominees, &nominees)

	return nominees
}

// GetPostPriceParams get nominee params from store
func (k Keeper) GetPostPriceParams(ctx sdk.Context) types.PostPriceParams {
	params := types.PostPriceParams{}
	k.paramstore.Get(ctx, types.KeyPostPrice, &params)

	return params
}

// GetPriceFeeds returns the price feeds in the price feed store
func (k Keeper) GetPriceFeeds(ctx sdk.Context, assetCode string) (types.PriceFeeds, error) {

	for _, a := range k.GetAssetParams(ctx) {
		if assetCode == a.AssetCode {
			return a.PriceFeeds, nil
		}
	}

	return types.PriceFeeds{}, fmt.Errorf("asset %q not found", assetCode)
}

// AddPriceFeed adds the price feed to the price feed store for specific assetCode
func (k Keeper) AddPriceFeed(ctx sdk.Context, nominee string, assetCode string, address sdk.AccAddress) error {
	if !k.IsNominee(ctx, nominee) {
		return fmt.Errorf("%q is not a valid nominee", nominee)
	}

	if _, err := k.GetPriceFeed(ctx, assetCode, address); err == nil {
		return fmt.Errorf("price feed %q already exists for asset %q", address, assetCode)
	}

	assets := k.GetAssetParams(ctx)
	updateAssets := assets[:0]
	found := false
	for _, a := range assets {
		if assetCode == a.AssetCode {
			pricefeed := types.NewOracle(address)
			a.PriceFeeds = append(a.PriceFeeds, pricefeed)
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

// SetPriceFeeds sets the price feed store for specific assetCode
func (k Keeper) SetPriceFeeds(ctx sdk.Context, nominee string, assetCode string, addresses types.PriceFeeds) error {
	if !k.IsNominee(ctx, nominee) {
		return fmt.Errorf("%q is not a valid nominee", nominee)
	}

	assets := k.GetAssetParams(ctx)
	updateAssets := assets[:0]
	found := false
	for _, a := range assets {
		if assetCode == a.AssetCode {
			a.PriceFeeds = addresses
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

// SetAsset overwrites existing asset for specific assetCode
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

// AddAsset adds non-existing asset to the store
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

// GetPriceFeed returns the price feed from the store or an error if not found for specific assetCode
func (k Keeper) GetPriceFeed(ctx sdk.Context, assetCode string, address sdk.AccAddress) (types.PriceFeed, error) {
	pricefeeds, err := k.GetPriceFeeds(ctx, assetCode)
	if err != nil {
		return types.PriceFeed{}, fmt.Errorf("asset %q not found", assetCode)
	}
	for _, o := range pricefeeds {
		if address.Equals(o.Address) {
			return o, nil
		}
	}
	return types.PriceFeed{}, fmt.Errorf("price feed %q not found for asset %q", address, assetCode)
}

// GetAsset returns the asset if it is in the price feed system
func (k Keeper) GetAsset(ctx sdk.Context, assetCode string) (types.Asset, bool) {
	assets := k.GetAssetParams(ctx)

	for i := range assets {
		if assets[i].AssetCode == assetCode {
			return assets[i], true
		}
	}
	return types.Asset{}, false

}
