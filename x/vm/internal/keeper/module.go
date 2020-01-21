package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"wings-blockchain/x/vm/internal/types"
	"wings-blockchain/x/vm/internal/types/vm_grpc"
)

func (keeper Keeper) storeModule(ctx sdk.Context, accessPath vm_grpc.VMAccessPath, code types.Contract) sdk.Error {
	store := ctx.KVStore(keeper.storeKey)
	moduleKey := types.MakePathKey(accessPath, types.VMModuleType)

	if store.Has(moduleKey) {
		return types.ErrModuleExists(types.DecodeAddress(accessPath.Address), accessPath.Path)
	}

	store.Set(moduleKey, code)
	return nil
}

func (keeper Keeper) hasModule(ctx sdk.Context, accessPath vm_grpc.VMAccessPath) bool {
	store := ctx.KVStore(keeper.storeKey)
	moduleKey := types.MakePathKey(accessPath, types.VMModuleType)

	return store.Has(moduleKey)
}
