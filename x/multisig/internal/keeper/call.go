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

	id := k.nextCallID(ctx)
	if id.GT(dnTypes.NewIDFromUint64(0)) {
		return id.Decr()
	}

	return id
}

// StoreCall sets call object.
func (k Keeper) StoreCall(ctx sdk.Context, call types.Call) {
	k.modulePerms.AutoCheck(types.PermWrite)

	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetCallKey(call.ID), k.cdc.MustMarshalBinaryBare(call))
}

// createCall updates lastCallId, sets uniqueID-callID match, stores a new call and adds call to the queue.
func (k Keeper) createCall(ctx sdk.Context, call types.Call) {
	store := ctx.KVStore(k.storeKey)

	store.Set(types.LastCallIdKey, k.cdc.MustMarshalBinaryLengthPrefixed(call.ID))
	store.Set(types.GetUniqueIDKey(call.UniqueID), k.cdc.MustMarshalBinaryLengthPrefixed(call.ID))
	k.StoreCall(ctx, call)
	k.addCallToQueue(ctx, call.ID, call.Height)

	ctx.EventManager().EmitEvent(types.NewCallSubmittedEvent(call))
}

// getCall returns call from the storage.
func (k Keeper) getCall(ctx sdk.Context, id dnTypes.ID) types.Call {
	store := ctx.KVStore(k.storeKey)

	var call types.Call
	bs := store.Get(types.GetCallKey(id))
	k.cdc.MustUnmarshalBinaryBare(bs, &call)

	return call
}

// nextCallId return next unique call object ID.
func (k Keeper) nextCallID(ctx sdk.Context) dnTypes.ID {
	store := ctx.KVStore(k.storeKey)

	if !store.Has(types.LastCallIdKey) {
		return dnTypes.NewIDFromUint64(0)
	}

	var id dnTypes.ID
	bz := store.Get(types.LastCallIdKey)
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &id)

	return id.Incr()
}
