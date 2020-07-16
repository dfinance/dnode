// Orders module keeper creates, stores and removes order objects.
// Keeper locks, unlock, transfers account funds on order posting, canceling, executing.
package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/supply"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/dfinance/dnode/helpers/perms"
	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/markets"
	"github.com/dfinance/dnode/x/orders/internal/types"
)

// Module keeper object.
type Keeper struct {
	cdc          *codec.Codec
	storeKey     sdk.StoreKey
	bankKeeper   bank.Keeper
	supplyKeeper supply.Keeper
	marketKeeper markets.Keeper
	modulePerms  perms.ModulePermissions
}

// PostOrder creates a new order object and locks account funds (coins).
func (k Keeper) PostOrder(
	ctx sdk.Context,
	owner sdk.AccAddress,
	assetCode dnTypes.AssetCode,
	direction types.Direction,
	price sdk.Uint,
	quantity sdk.Uint,
	ttlInSec uint64) (types.Order, error) {

	k.modulePerms.AutoCheck(types.PermOrderPost)

	filter := markets.NewMarketsFilter(1, 1)
	filter.AssetCode = assetCode.String()

	marketsList := k.marketKeeper.GetListFiltered(ctx, filter)

	if len(marketsList) == 0 {
		return types.Order{}, sdkErrors.Wrap(types.ErrWrongAssetCode, "not found")
	}

	market, err := k.marketKeeper.GetExtended(ctx, marketsList[0].ID)
	if err != nil {
		return types.Order{}, err
	}

	id := k.nextID(ctx)
	order := types.NewOrder(ctx, id, owner, market, direction, price, quantity, ttlInSec)
	if err := order.ValidatePriceQuantity(); err != nil {
		return types.Order{}, err
	}

	if err := k.LockOrderCoins(ctx, order); err != nil {
		return types.Order{}, err
	}
	k.set(ctx, order)
	k.setID(ctx, id)

	ctx.EventManager().EmitEvent(types.NewOrderPostedEvent(order))

	k.GetLogger(ctx).Debug(fmt.Sprintf("order %s from %s: posted", id, owner))

	return order, nil
}

// RevokeOrder removes an order object and unlocks account funds (coins).
func (k Keeper) RevokeOrder(ctx sdk.Context, id dnTypes.ID) error {
	k.modulePerms.AutoCheck(types.PermOrderRevoke)

	order, err := k.Get(ctx, id)
	if err != nil {
		return sdkErrors.Wrap(types.ErrWrongOrderID, "not found")
	}

	if err := k.UnlockOrderCoins(ctx, order); err != nil {
		return err
	}
	k.del(ctx, id)

	ctx.EventManager().EmitEvent(types.NewOrderCanceledEvent(order))

	return nil
}

// GetLogger gets logger with keeper context.
func (k Keeper) GetLogger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// nextID return next unique order object ID.
func (k Keeper) nextID(ctx sdk.Context) dnTypes.ID {
	store := ctx.KVStore(k.storeKey)
	if !store.Has(types.LastOrderIDKey) {
		return dnTypes.NewIDFromUint64(0)
	}

	bz := store.Get(types.LastOrderIDKey)
	id := dnTypes.ID{}
	if err := k.cdc.UnmarshalBinaryLengthPrefixed(bz, &id); err != nil {
		panic(fmt.Errorf("lastOrderID unmarshal: %w", err))
	}

	return id.Incr()
}

// setID sets last unique order object ID.
func (k Keeper) setID(ctx sdk.Context, id dnTypes.ID) {
	store := ctx.KVStore(k.storeKey)

	bz, err := k.cdc.MarshalBinaryLengthPrefixed(id)
	if err != nil {
		panic(fmt.Errorf("lastOrderID marshal: %w", err))
	}

	store.Set(types.LastOrderIDKey, bz)
}

// NewKeeper creates keeper object.
func NewKeeper(
	cdc *codec.Codec,
	storeKey sdk.StoreKey,
	bk bank.Keeper,
	sk supply.Keeper,
	mk markets.Keeper,
	permsRequesters ...perms.RequestModulePermissions,
) Keeper {
	k := Keeper{
		cdc:          cdc,
		storeKey:     storeKey,
		bankKeeper:   bk,
		supplyKeeper: sk,
		marketKeeper: mk,
		modulePerms:  types.NewModulePerms(),
	}
	for _, requester := range permsRequesters {
		k.modulePerms.AutoAddRequester(requester)
	}

	return k
}
