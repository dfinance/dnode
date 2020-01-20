package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	vm "wings-blockchain/x/core/protos"
	"wings-blockchain/x/vm/internal/types"
)

func (keeper Keeper) storeModule(ctx sdk.Context, accessPath vm.VMAccessPath, code types.Contract) sdk.Error {
	store := ctx.KVStore(keeper.storeKey)
	moduleKey := types.MakePathKey(accessPath, types.VMModuleType)

	if store.Has(moduleKey) {
		return types.ErrModuleExists(types.DecodeAddress(accessPath.Address), accessPath.Path)
	}

	store.Set(moduleKey, code)
	return nil
}
