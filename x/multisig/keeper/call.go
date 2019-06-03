package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"wings-blockchain/x/multisig/types"
)

// Submit call to execute by confirmations from validators
func (keeper Keeper) SubmitCall(ctx sdk.Context, msg types.MsMsg) sdk.Error {
	if !keeper.router.HasRoute(msg.Route()) {
		return types.ErrRouteDoesntExist(msg.Route())
	}

	nextId := keeper.getNextCallId(ctx)
	call   := types.NewCall(msg)
	keeper.saveCallById(ctx, nextId, call)

	return nil
}

func (keeper Keeper) GetCall(ctx sdk.Context, id uint64) (sdk.Error, types.Call) {
	if !keeper.HasCall(ctx, id) {
		return types.ErrWrongCallId(id), types.Call{}
	}

	return nil, keeper.getCallById(ctx, id)
}

// Check if call exists
func (keeper Keeper) HasCall(ctx sdk.Context, id uint64) bool {
	store := ctx.KVStore(keeper.storeKey)

	return store.Has(types.GetCallByIdKey(id))
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


