package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/currencies/internal/types"
)

// DestroyCurrency lowers payee coin balance.
func (k Keeper) DestroyCurrency(ctx sdk.Context, denom string, amount sdk.Int, spender sdk.AccAddress, recipient, chainID string) error {
	if !k.HasCurrency(ctx, denom) {
		return sdkErrors.Wrapf(sdkErrors.ErrInsufficientFunds, "denom %q: not found", denom)
	}

	k.reduceSupply(ctx, denom, amount, spender, recipient, chainID)

	newCoin := sdk.NewCoin(denom, amount)
	if _, err := k.bankKeeper.SubtractCoins(ctx, spender, sdk.Coins{newCoin}); err != nil {
		return err
	}

	return nil
}

// HasDestroy checks that destroy exists.
func (k Keeper) HasDestroy(ctx sdk.Context, id dnTypes.ID) bool {
	store := ctx.KVStore(k.storeKey)

	return store.Has(types.GetDestroyKey(id))
}

// GetDestroy returns destroy.
func (k Keeper) GetDestroy(ctx sdk.Context, id dnTypes.ID) (types.Destroy, error) {
	if !k.HasDestroy(ctx, id) {
		return types.Destroy{}, sdkErrors.Wrapf(types.ErrWrongDestroyID, "destroyID %q: not found", id.String())
	}

	return k.getDestroy(ctx, id), nil
}

// GetDestroysFiltered returns destroy objects list with pagination params.
func (k Keeper) GetDestroysFiltered(ctx sdk.Context, params types.DestroysReq) (types.Destroys, error) {
	if params.Page.GT(sdk.ZeroUint()) {
		params.Page = params.Page.SubUint64(1)
	}

	startID := params.Page.Mul(params.Limit)
	endID := startID.Add(params.Limit)

	destroys := make(types.Destroys, 0)
	for ; startID.LT(endID); startID = startID.AddUint64(1) {
		id, _ := dnTypes.NewIDFromString(startID.String())
		destroy, err := k.GetDestroy(ctx, id)
		if err != nil {
			break
		}

		destroys = append(destroys, destroy)
	}

	return destroys, nil
}

// getDestroy returns destroy from the storage.
func (k Keeper) getDestroy(ctx sdk.Context, id dnTypes.ID) types.Destroy {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetDestroyKey(id))

	destroy := types.Destroy{}
	k.cdc.MustUnmarshalBinaryBare(bz, &destroy)

	return destroy
}

// storeDestroy sets destroy to the storage.
func (k Keeper) storeDestroy(ctx sdk.Context, destroy types.Destroy) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetDestroyKey(destroy.ID), k.cdc.MustMarshalBinaryBare(destroy))
}

// setLastDestroyID sets lastDestroyID to the storage.
func (k Keeper) setLastDestroyID(ctx sdk.Context, id dnTypes.ID) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetLastDestroyIDKey(), k.cdc.MustMarshalBinaryBare(id))
}

// getLastDestroyID gets lastDestroyID from the storage.
func (k Keeper) getLastDestroyID(ctx sdk.Context) dnTypes.ID {
	store := ctx.KVStore(k.storeKey)

	id := dnTypes.ID{}
	k.cdc.MustUnmarshalBinaryBare(store.Get(types.GetLastDestroyIDKey()), &id)

	return id
}

// getNewDestroyID creates next lastDestroyID.
func (k Keeper) getNextDestroyID(ctx sdk.Context) dnTypes.ID {
	store := ctx.KVStore(k.storeKey)
	if !store.Has(types.GetLastDestroyIDKey()) {
		return dnTypes.NewIDFromUint64(0)
	}
	lastID := k.getLastDestroyID(ctx)

	return lastID.Incr()
}
