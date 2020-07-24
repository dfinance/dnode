package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/dfinance/dnode/helpers"
	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/markets/internal/types"
)

// Has check if market object with ID exists.
func (k Keeper) Has(ctx sdk.Context, id dnTypes.ID) bool {
	k.modulePerms.AutoCheck(types.PermRead)

	store := ctx.KVStore(k.storeKey)

	return store.Has(types.GetMarketsKey(id))
}

// Get gets market object by ID.
func (k Keeper) Get(ctx sdk.Context, id dnTypes.ID) (types.Market, error) {
	k.modulePerms.AutoCheck(types.PermRead)

	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetMarketsKey(id))
	if bz == nil {
		return types.Market{}, types.ErrWrongID
	}

	market := types.Market{}
	if err := k.cdc.UnmarshalBinaryLengthPrefixed(bz, &market); err != nil {
		panic(fmt.Errorf("market unmarshal: %w", err))
	}

	return market, nil
}

// GetExtended gets currency infos and build a MarketExtended object.
func (k Keeper) GetExtended(ctx sdk.Context, id dnTypes.ID) (retMarket types.MarketExtended, retErr error) {
	k.modulePerms.AutoCheck(types.PermRead)

	market, err := k.Get(ctx, id)
	if err != nil {
		retErr = err
		return
	}

	baseCurrency, err := k.ccsStorage.GetCurrency(ctx, market.BaseAssetDenom)
	if err != nil {
		retErr = sdkErrors.Wrap(err, "BaseAsset")
	}

	quoteCurrency, err := k.ccsStorage.GetCurrency(ctx, market.QuoteAssetDenom)
	if err != nil {
		retErr = sdkErrors.Wrap(err, "QuoteAsset")
	}

	retMarket = types.NewMarketExtended(market, baseCurrency, quoteCurrency)

	return
}

// Add creates a new market object.
// Action is only allowed to nominee accounts.
func (k Keeper) Add(ctx sdk.Context, baseAsset, quoteAsset string) (types.Market, error) {
	k.modulePerms.AutoCheck(types.PermCreate)

	// check if market already exists
	var duplicatedErr error
	k.iterateMarkets(ctx, func(m types.Market) bool {
		if m.BaseAssetDenom == baseAsset && m.QuoteAssetDenom == quoteAsset {
			duplicatedErr = sdkErrors.Wrap(types.ErrMarketExists, m.String())
			return false
		}
		return true
	})
	if duplicatedErr != nil {
		return types.Market{}, duplicatedErr
	}

	// check currencies do exist
	if !k.ccsStorage.HasCurrency(ctx, baseAsset) {
		return types.Market{}, sdkErrors.Wrap(types.ErrWrongAssetDenom, "BaseAsset not registered")
	}
	if !k.ccsStorage.HasCurrency(ctx, quoteAsset) {
		return types.Market{}, sdkErrors.Wrap(types.ErrWrongAssetDenom, "QuoteAsset not registered")
	}

	market := types.NewMarket(k.nextID(ctx), baseAsset, quoteAsset)
	k.set(ctx, market)
	k.setLastID(ctx, market.ID)

	ctx.EventManager().EmitEvent(types.NewMarketCreatedEvent(market))

	return market, nil
}

// GetList returns all market objects.
func (k Keeper) GetList(ctx sdk.Context) types.Markets {
	k.modulePerms.AutoCheck(types.PermRead)

	markets := make(types.Markets, 0)

	k.iterateMarkets(ctx, func(m types.Market) bool {
		markets = append(markets, m)
		return true
	})

	return markets
}

// GetListFiltered returns market objects filtered by params.
func (k Keeper) GetListFiltered(ctx sdk.Context, params types.MarketsReq) types.Markets {
	k.modulePerms.AutoCheck(types.PermRead)

	filteredMarkets := make(types.Markets, 0)

	k.iterateMarkets(ctx, func(m types.Market) bool {
		add := true

		if params.BaseDenomFilter() && m.BaseAssetDenom != params.BaseAssetDenom {
			add = false
		}
		if params.QuoteDenomFilter() && m.QuoteAssetDenom != params.QuoteAssetDenom {
			add = false
		}
		if params.AssetCodeFilter() && params.AssetCode != m.GetAssetCode().String() {
			add = false
		}

		if add {
			filteredMarkets = append(filteredMarkets, m)
		}

		return true
	})

	start, end, err := helpers.PaginateSlice(len(filteredMarkets), params.Page, params.Limit)
	if err != nil {
		return types.Markets{}
	}

	return filteredMarkets[start:end]
}

// set creates / overwrites market object in the storage.
func (k Keeper) set(ctx sdk.Context, market types.Market) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetMarketsKey(market.ID)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(market)
	store.Set(key, bz)
}

func (k Keeper) setLastID(ctx sdk.Context, id dnTypes.ID) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(id)
	store.Set(types.KeyLastMarketId, bz)
}

// getLastMarketID returns lastMarketID from the storage if exists.
func (k Keeper) getLastMarketID(ctx sdk.Context) *dnTypes.ID {
	store := ctx.KVStore(k.storeKey)

	if !store.Has(types.KeyLastMarketId) {
		return nil
	}

	var id dnTypes.ID
	bz := store.Get(types.KeyLastMarketId)
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &id)

	return &id
}

// nextID return next unique market object ID.
func (k Keeper) nextID(ctx sdk.Context) dnTypes.ID {
	id := k.getLastMarketID(ctx)
	if id == nil {
		return dnTypes.NewZeroID()
	}

	return id.Incr()
}

// iterateMarkets iterates through all registered markets and execs handler on each.
func (k Keeper) iterateMarkets(ctx sdk.Context, handler func(market types.Market) bool) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.GetPrefixMarketsKey())
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var market types.Market
		k.cdc.MustUnmarshalBinaryLengthPrefixed(iterator.Value(), &market)
		if !handler(market) {
			break
		}
	}
}
