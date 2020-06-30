package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/currencies/internal/types"
)

// WithdrawCurrency lowers payee coin balance.
func (k Keeper) WithdrawCurrency(ctx sdk.Context, denom string, amount sdk.Int, spender sdk.AccAddress, recipient, chainID string) (retErr error) {
	// bankKeeper might panic
	defer func() {
		if r := recover(); r != nil {
			retErr = sdkErrors.Wrapf(types.ErrInternal, "bankKeeper.SubtractCoins for address %q panic: %v", spender, r)
		}
	}()

	if !k.HasCurrency(ctx, denom) {
		return sdkErrors.Wrapf(sdkErrors.ErrInsufficientFunds, "denom %q: not found", denom)
	}

	k.reduceSupply(ctx, denom, amount, spender, recipient, chainID)

	newCoin := sdk.NewCoin(denom, amount)
	if _, err := k.bankKeeper.SubtractCoins(ctx, spender, sdk.Coins{newCoin}); err != nil {
		return err
	}

	return
}

// HasWithdraw checks that withdraw exists.
func (k Keeper) HasWithdraw(ctx sdk.Context, id dnTypes.ID) bool {
	store := ctx.KVStore(k.storeKey)

	return store.Has(types.GetWithdrawKey(id))
}

// GetWithdraw returns withdraw.
func (k Keeper) GetWithdraw(ctx sdk.Context, id dnTypes.ID) (types.Withdraw, error) {
	if !k.HasWithdraw(ctx, id) {
		return types.Withdraw{}, sdkErrors.Wrapf(types.ErrWrongWithdrawID, "withdrawID %q: not found", id.String())
	}

	return k.getWithdraw(ctx, id), nil
}

// GetWithdrawsFiltered returns withdraw objects list with pagination params.
func (k Keeper) GetWithdrawsFiltered(ctx sdk.Context, params types.WithdrawsReq) (types.Withdraws, error) {
	if params.Page.GT(sdk.ZeroUint()) {
		params.Page = params.Page.SubUint64(1)
	}

	startID := params.Page.Mul(params.Limit)
	endID := startID.Add(params.Limit)

	withdraws := make(types.Withdraws, 0)
	for ; startID.LT(endID); startID = startID.AddUint64(1) {
		id, _ := dnTypes.NewIDFromString(startID.String())
		withdraw, err := k.GetWithdraw(ctx, id)
		if err != nil {
			break
		}

		withdraws = append(withdraws, withdraw)
	}

	return withdraws, nil
}

// getWithdraw returns withdraw from the storage.
func (k Keeper) getWithdraw(ctx sdk.Context, id dnTypes.ID) types.Withdraw {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetWithdrawKey(id))

	withdraw := types.Withdraw{}
	k.cdc.MustUnmarshalBinaryBare(bz, &withdraw)

	return withdraw
}

// storeWithdraw sets withdraw to the storage.
func (k Keeper) storeWithdraw(ctx sdk.Context, withdraw types.Withdraw) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetWithdrawKey(withdraw.ID), k.cdc.MustMarshalBinaryBare(withdraw))
}

// setLastWithdrawID sets lastWithdrawID to the storage.
func (k Keeper) setLastWithdrawID(ctx sdk.Context, id dnTypes.ID) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetLastWithdrawIDKey(), k.cdc.MustMarshalBinaryBare(id))
}

// getLastWithdrawID gets lastWithdrawID from the storage.
func (k Keeper) getLastWithdrawID(ctx sdk.Context) dnTypes.ID {
	store := ctx.KVStore(k.storeKey)

	id := dnTypes.ID{}
	k.cdc.MustUnmarshalBinaryBare(store.Get(types.GetLastWithdrawIDKey()), &id)

	return id
}

// getNextWithdrawID creates next lastWithdrawID.
func (k Keeper) getNextWithdrawID(ctx sdk.Context) dnTypes.ID {
	store := ctx.KVStore(k.storeKey)
	if !store.Has(types.GetLastWithdrawIDKey()) {
		return dnTypes.NewIDFromUint64(0)
	}
	lastID := k.getLastWithdrawID(ctx)

	return lastID.Incr()
}
