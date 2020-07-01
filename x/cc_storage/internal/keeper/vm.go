package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/dfinance/dnode/x/cc_storage/internal/types"
)

// GetCurrencyBalancePath returns VM balance path for currency.
func (k Keeper) GetCurrencyBalancePath(ctx sdk.Context, denom string) ([]byte, error) {
	path, ok := k.getPathData(ctx, types.GetCurrencyBalancePathKey(denom))
	if !ok {
		return nil, sdkErrors.Wrapf(types.ErrWrongDenom, "balancePath for %q currency: not found", denom)
	}

	return path, nil
}

// GetCurrencyInfoPath returns VM info path for currency.
func (k Keeper) GetCurrencyInfoPath(ctx sdk.Context, denom string) ([]byte, error) {
	path, ok := k.getPathData(ctx, types.GetCurrencyInfoPathKey(denom))
	if !ok {
		return nil, sdkErrors.Wrapf(types.ErrWrongDenom, "infoPath for %q currency: not found", denom)
	}

	return path, nil
}

// storeCurrencyBalancePath sets currency balance VM path to the params storage.
func (k Keeper) storeCurrencyBalancePath(ctx sdk.Context, denom string, path []byte) {
	k.storePathData(ctx, types.GetCurrencyBalancePathKey(denom), path)
}

// storeCurrencyInfoPath sets currency info VM path to the params storage.
func (k Keeper) storeCurrencyInfoPath(ctx sdk.Context, denom string, path []byte) {
	k.storePathData(ctx, types.GetCurrencyInfoPathKey(denom), path)
}

// getPathData gets PathData from the storage.
func (k Keeper) getPathData(ctx sdk.Context, key []byte) ([]byte, bool) {
	storage := ctx.KVStore(k.storeKey)
	if !storage.Has(key) {
		return nil, false
	}

	data := types.PathData{}
	bz := storage.Get(key)
	if err := k.cdc.UnmarshalBinaryBare(bz, &data); err != nil {
		panic(fmt.Errorf("unmarshal PathData: %v", err))
	}

	return data.Path, true
}

// storePathData stored PathData to the storage.
func (k Keeper) storePathData(ctx sdk.Context, key, path []byte) {
	storage := ctx.KVStore(k.storeKey)

	data := types.PathData{Path: path}
	bz, err := k.cdc.MarshalBinaryBare(data)
	if err != nil {
		panic(fmt.Errorf("marshal PathData: %v", err))
	}
	storage.Set(key, bz)
}
