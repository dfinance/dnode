package keeper

import (
	"fmt"
	dnTypes "github.com/dfinance/dnode/helpers/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dfinance/dnode/x/oracle/internal/types"
)

// GetOracles returns oracles.
func (k Keeper) GetOracles(ctx sdk.Context, assetCode dnTypes.AssetCode) (types.Oracles, error) {
	for _, a := range k.GetAssetParams(ctx) {
		if assetCode == a.AssetCode {
			return a.Oracles, nil
		}
	}

	return types.Oracles{}, fmt.Errorf("oracles for %q: not found", assetCode)
}

// GetOracle returns an oracle for specific assetCode.
func (k Keeper) GetOracle(ctx sdk.Context, assetCode dnTypes.AssetCode, address sdk.AccAddress) (types.Oracle, error) {
	oracles, err := k.GetOracles(ctx, assetCode)
	if err != nil {
		return types.Oracle{}, err
	}

	for _, o := range oracles {
		if address.Equals(o.Address) {
			return o, nil
		}
	}

	return types.Oracle{}, fmt.Errorf("oracle %q for asset %q: not found", address, assetCode)
}

// AddOracle adds an oracle to specific assetCode.
func (k Keeper) AddOracle(ctx sdk.Context, nominee string, assetCode dnTypes.AssetCode, address sdk.AccAddress) error {
	if err := k.IsNominee(ctx, nominee); err != nil {
		return err
	}

	if _, err := k.GetOracle(ctx, assetCode, address); err == nil {
		return fmt.Errorf("oracle %q for asset %q: already exists", address, assetCode)
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

	return fmt.Errorf("asset %q: not found", assetCode)
}

// SetOracles sets (overwrites) oracles for specific assetCode.
func (k Keeper) SetOracles(ctx sdk.Context, nominee string, assetCode dnTypes.AssetCode, addresses types.Oracles) error {
	if err := k.IsNominee(ctx, nominee); err != nil {
		return err
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

	return fmt.Errorf("asset %q: not found", assetCode)
}
