package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/core/msmodule"
	"github.com/dfinance/dnode/x/multisig/internal/types"
)

// SubmitCall creates a new call to be executed on validators by confirmation.
func (k Keeper) SubmitCall(ctx sdk.Context, msg msmodule.MsMsg, uniqueID string, sender sdk.AccAddress) error {
	k.modulePerms.AutoCheck(types.PermWrite)

	// check call exists
	if k.HasCallByUniqueID(ctx, uniqueID) {
		return sdkErrors.Wrapf(types.ErrWrongCallUniqueId, "%q exists", uniqueID)
	}

	// create a new call and check its validity
	call, err := types.NewCall(k.nextCallID(ctx), uniqueID, msg, ctx.BlockHeight(), sender)
	if err != nil {
		return err
	}

	// check msg route and dry-run handler
	if !k.router.HasRoute(msg.Route()) {
		return sdkErrors.Wrapf(types.ErrWrongMsgRoute, "%q not found", msg.Route())
	}

	cacheCtx, _ := ctx.CacheContext()
	handler := k.router.GetRoute(msg.Route())
	if err := handler(cacheCtx, msg); err != nil {
		return err
	}

	// create call and confirm it by its creator
	k.createCall(ctx, call)
	if err := k.ConfirmCall(ctx, call.ID, sender); err != nil {
		return err
	}

	return nil
}

// HasCall checks that call exists.
func (k Keeper) HasCall(ctx sdk.Context, id dnTypes.ID) bool {
	k.modulePerms.AutoCheck(types.PermRead)

	store := ctx.KVStore(k.storeKey)

	return store.Has(types.GetCallKey(id))
}

// HasCallByUniqueID checks that call with uniqueID exists.
func (k Keeper) HasCallByUniqueID(ctx sdk.Context, uniqueID string) bool {
	k.modulePerms.AutoCheck(types.PermRead)

	store := ctx.KVStore(k.storeKey)

	return store.Has(types.GetUniqueIDKey(uniqueID))
}

// GetCall returns call.
func (k Keeper) GetCall(ctx sdk.Context, id dnTypes.ID) (types.Call, error) {
	k.modulePerms.AutoCheck(types.PermRead)

	if !k.HasCall(ctx, id) {
		return types.Call{}, sdkErrors.Wrapf(types.ErrWrongCallId, "%s not found", id.String())
	}

	return k.getCall(ctx, id), nil
}

// GetCallIDByUnique return callID by its uniqueID.
func (k Keeper) GetCallIDByUniqueID(ctx sdk.Context, uniqueID string) (dnTypes.ID, error) {
	k.modulePerms.AutoCheck(types.PermRead)

	store := ctx.KVStore(k.storeKey)

	if !k.HasCallByUniqueID(ctx, uniqueID) {
		return dnTypes.ID{}, sdkErrors.Wrapf(types.ErrWrongCallUniqueId, "%q not found", uniqueID)
	}
	bz := store.Get(types.GetUniqueIDKey(uniqueID))

	var id dnTypes.ID
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &id)

	return id, nil
}

// GetLastID returns last created call ID.
func (k Keeper) GetLastCallID(ctx sdk.Context) dnTypes.ID {
	k.modulePerms.AutoCheck(types.PermRead)

	store := ctx.KVStore(k.storeKey)

	if !store.Has(types.LastCallIdKey) {
		return dnTypes.NewIDFromUint64(0)
	}

	var id dnTypes.ID
	bz := store.Get(types.LastCallIdKey)
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &id)

	return id
}

// StoreCall sets call object.
func (k Keeper) StoreCall(ctx sdk.Context, call types.Call) {
	k.modulePerms.AutoCheck(types.PermWrite)

	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetCallKey(call.ID), k.cdc.MustMarshalBinaryBare(call))
}

// getCall returns call from the storage.
func (k Keeper) getCall(ctx sdk.Context, id dnTypes.ID) types.Call {
	store := ctx.KVStore(k.storeKey)

	var call types.Call
	bs := store.Get(types.GetCallKey(id))
	k.cdc.MustUnmarshalBinaryBare(bs, &call)

	return call
}

// getCalls returns all registered call objects.
func (k Keeper) getCalls(ctx sdk.Context) []types.Call {
	calls := make([]types.Call, 0)

	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.GetCallKeyPrefix())
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var call types.Call
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &call)
		calls = append(calls, call)
	}

	return calls
}

// createCall updates lastCallId, sets uniqueID-callID match, stores a new call and adds call to the queue.
func (k Keeper) createCall(ctx sdk.Context, call types.Call) {
	k.setLastCallID(ctx, call.ID)
	k.setCallUniqueIDMatch(ctx, call.UniqueID, call.ID)
	k.StoreCall(ctx, call)
	k.addCallToQueue(ctx, call.ID, call.Height)

	ctx.EventManager().EmitEvent(types.NewCallSubmittedEvent(call))
}

// setCallUniqueIDMatch
func (k Keeper) setCallUniqueIDMatch(ctx sdk.Context, uniqueID string, callID dnTypes.ID) {
	store := ctx.KVStore(k.storeKey)

	store.Set(types.GetUniqueIDKey(uniqueID), k.cdc.MustMarshalBinaryLengthPrefixed(callID))
}

// nextCallId return next unique call object ID (first is 0).
func (k Keeper) nextCallID(ctx sdk.Context) dnTypes.ID {
	id := k.getLastCallID(ctx)
	if id == nil {
		return dnTypes.NewZeroID()
	}

	return id.Incr()
}

// getLastCallID returns lastCallID from the storage if exists.
func (k Keeper) getLastCallID(ctx sdk.Context) *dnTypes.ID {
	store := ctx.KVStore(k.storeKey)

	if !store.Has(types.LastCallIdKey) {
		return nil
	}

	var id dnTypes.ID
	bz := store.Get(types.LastCallIdKey)
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &id)

	return &id
}

// setLastCallID sets lastCallID to the storage.
func (k Keeper) setLastCallID(ctx sdk.Context, id dnTypes.ID) {
	store := ctx.KVStore(k.storeKey)

	store.Set(types.LastCallIdKey, k.cdc.MustMarshalBinaryLengthPrefixed(id))
}
