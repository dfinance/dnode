package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dfinance/dnode/x/oracle/internal/types"
)

// GetOracles returns the oracles in the oracle store.
func (k Keeper) GetOracles(ctx sdk.Context, assetCode string) (types.Oracles, error) {

	for _, a := range k.GetAssetParams(ctx) {
		if assetCode == a.AssetCode {
			return a.Oracles, nil
		}
	}

	return types.Oracles{}, fmt.Errorf("asset %q not found", assetCode)
}

// GetOracle returns the oracle from the store or an error if not found for specific assetCode.
func (k Keeper) GetOracle(ctx sdk.Context, assetCode string, address sdk.AccAddress) (types.Oracle, error) {
	oracles, err := k.GetOracles(ctx, assetCode)
	if err != nil {
		return types.Oracle{}, fmt.Errorf("asset %q not found", assetCode)
	}
	for _, o := range oracles {
		if address.Equals(o.Address) {
			return o, nil
		}
	}
	return types.Oracle{}, fmt.Errorf("oracle %q not found for asset %q", address, assetCode)
}

// AddOracle adds the oracle to the oracle store for specific assetCode.
func (k Keeper) AddOracle(ctx sdk.Context, nominee string, assetCode string, address sdk.AccAddress) error {
	if !k.IsNominee(ctx, nominee) {
		return fmt.Errorf("%q is not a valid nominee", nominee)
	}

	if _, err := k.GetOracle(ctx, assetCode, address); err == nil {
		return fmt.Errorf("oracle %q already exists for asset %q", address, assetCode)
	}

	assets := k.GetAssetParams(ctx)
	updateAssets := assets[:0]
	found := false
	for _, a := range assets {
		if assetCode == a.AssetCode {
			oracle := types.NewOracle(address)
			a.Oracles = append(a.Oracles, oracle)
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

// SetOracles sets the oracle store for specific assetCode.
func (k Keeper) SetOracles(ctx sdk.Context, nominee string, assetCode string, addresses types.Oracles) error {
	if !k.IsNominee(ctx, nominee) {
		return fmt.Errorf("%q is not a valid nominee", nominee)
	}

	assets := k.GetAssetParams(ctx)
	updateAssets := assets[:0]
	found := false
	for _, a := range assets {
		if assetCode == a.AssetCode {
			a.Oracles = addresses
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
