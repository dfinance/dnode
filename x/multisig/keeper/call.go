package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"wings-blockchain/x/multisig/types"
)

// Submit call to execute by confirmations from validators
func (keeper Keeper) SubmitCall(ctx sdk.Context, msg types.MsMsg, uniqueID string, sender sdk.AccAddress) sdk.Error {
	if !keeper.router.HasRoute(msg.Route()) {
		return types.ErrRouteDoesntExist(msg.Route())
	}

	cacheCtx, _ := ctx.CacheContext()
	handler := keeper.router.GetRoute(msg.Route())

	err := handler(cacheCtx, msg)

	if err != nil {
		return err
	}

	if keeper.HasCallByUniqueId(ctx, uniqueID) {
	    return types.ErrNotUniqueID(uniqueID)
    }

	nextId := keeper.getNextCallId(ctx)
	call, err := types.NewCall(nextId, uniqueID, msg, ctx.BlockHeight(), sender)

	if err != nil {
        return err
    }

	id := keeper.saveNewCall(ctx, call)

	keeper.addCallToQueue(ctx, id, call.Height)

	err = keeper.Confirm(ctx, id, sender)

	if err != nil {
		return err
	}

	return nil
}

// Get call by id
func (keeper Keeper) GetCall(ctx sdk.Context, id uint64) (types.Call, sdk.Error) {
	if !keeper.HasCall(ctx, id) {
		return types.Call{}, types.ErrWrongCallId(id)
	}

	return keeper.getCallById(ctx, id), nil
}

// Get call by unique id
func (keeper Keeper) GetCallIDByUnique(ctx sdk.Context, uniqueID string) (uint64, sdk.Error) {
    store := ctx.KVStore(keeper.storeKey)

    if !keeper.HasCallByUniqueId(ctx, uniqueID) {
        return 0, types.ErrNotFoundUniqueID(uniqueID)
    }

    bz := store.Get(types.GetUniqueID(uniqueID))

    var id uint64
    keeper.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &id)

    return id, nil
}

// Check if call exists
func (keeper Keeper) HasCall(ctx sdk.Context, id uint64) bool {
	store := ctx.KVStore(keeper.storeKey)

	return store.Has(types.GetCallByIdKey(id))
}

// Check if has call by unique id
func (keeper Keeper) HasCallByUniqueId(ctx sdk.Context, uniqueID string) bool {
    store := ctx.KVStore(keeper.storeKey)

    return store.Has(types.GetUniqueID(uniqueID))
}

// Get last call id
func (keeper Keeper) GetLastId(ctx sdk.Context) uint64  {
	id := keeper.getNextCallId(ctx)

	if id == 0 {
		return id
	}

	return id-1
}

// Save new call
func (keeper Keeper) saveNewCall(ctx sdk.Context, call types.Call) uint64 {
	store  := ctx.KVStore(keeper.storeKey)
	nextId := keeper.getNextCallId(ctx)

	store.Set(types.GetCallByIdKey(nextId), keeper.cdc.MustMarshalBinaryBare(call))
	store.Set(types.GetUniqueID(call.UniqueID), keeper.cdc.MustMarshalBinaryLengthPrefixed(nextId))
	store.Set(types.LastCallId, keeper.cdc.MustMarshalBinaryLengthPrefixed(nextId))

	return nextId
}

// Save message by id
func (keeper Keeper) saveCallById(ctx sdk.Context, id uint64, call types.Call) {
	store := ctx.KVStore(keeper.storeKey)

	store.Set(types.GetCallByIdKey(id), keeper.cdc.MustMarshalBinaryBare(call))
}

// Get message by id
func (keeper Keeper) getCallById(ctx sdk.Context, id uint64) types.Call {
	store := ctx.KVStore(keeper.storeKey)

	var call types.Call
	bs := store.Get(types.GetCallByIdKey(id))

	keeper.cdc.MustUnmarshalBinaryBare(bs, &call)
	return call
}

// Get next id to store message
func (keeper Keeper) getNextCallId(ctx sdk.Context) uint64 {
	store := ctx.KVStore(keeper.storeKey)

	if !store.Has(types.LastCallId) {
		return 0
	}

	b := store.Get(types.LastCallId)

	var id uint64
	err := keeper.cdc.UnmarshalBinaryLengthPrefixed(b, &id)

	if err != nil {
		panic(err)
	}

	id += 1

	return id
}


