package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/supply"
	"github.com/tendermint/tendermint/libs/log"

	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/market"
	"github.com/dfinance/dnode/x/order/internal/types"
)

type Keeper struct {
	cdc          *codec.Codec
	storeKey     sdk.StoreKey
	bankKeeper   bank.Keeper
	supplyKeeper supply.Keeper
	marketKeeper market.Keeper
}

func NewKeeper(storeKey sdk.StoreKey, cdc *codec.Codec, bk bank.Keeper, sk supply.Keeper, mk market.Keeper) Keeper {
	return Keeper{
		cdc:          cdc,
		storeKey:     storeKey,
		bankKeeper:   bk,
		supplyKeeper: sk,
		marketKeeper: mk,
	}
}

func (k Keeper) PostOrder(
	ctx sdk.Context,
	owner sdk.AccAddress,
	marketID dnTypes.ID,
	direction types.Direction,
	price sdk.Uint,
	quantity sdk.Uint,
	ttlInSec uint64) (types.Order, error) {

	market, err := k.marketKeeper.Get(ctx, marketID)
	if err != nil {
		return types.Order{}, sdkErrors.Wrap(types.ErrWrongMarketID, "not found")
	}

	id := k.nextID(ctx)
	order := types.NewOrder(ctx, id, owner, market, direction, price, quantity, ttlInSec)

	if err := k.LockOrderCoins(ctx, order); err != nil {
		return types.Order{}, err
	}
	k.Set(ctx, order)
	k.setID(ctx, id)

	return order, nil
}

func (k Keeper) CancelOrder(ctx sdk.Context, id dnTypes.ID) error {
	order, err := k.Get(ctx, id)
	if err != nil {
		return sdkErrors.Wrap(types.ErrWrongOrderID, "not found")
	}

	if err := k.UnlockOrderCoins(ctx, order); err != nil {
		return err
	}
	k.Del(ctx, id)

	return nil
}

// Get logger for keeper.
func (k Keeper) GetLogger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/" + types.ModuleName)
}

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

func (k Keeper) setID(ctx sdk.Context, id dnTypes.ID) {
	store := ctx.KVStore(k.storeKey)

	bz, err := k.cdc.MarshalBinaryLengthPrefixed(id)
	if err != nil {
		panic(fmt.Errorf("lastOrderID marshal: %w", err))
	}

	store.Set(types.LastOrderIDKey, bz)
}
