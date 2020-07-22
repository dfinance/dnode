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

	_, err := k.Get(ctx, id)

	return err == nil
}

// Get gets market object by ID.
func (k Keeper) Get(ctx sdk.Context, id dnTypes.ID) (types.Market, error) {
	k.modulePerms.AutoCheck(types.PermRead)

	params := k.GetParams(ctx)
	nextID := k.nextID(params)

	if !id.LT(nextID) {
		return types.Market{}, types.ErrWrongID
	}

	market := params.Markets[id.UInt64()]
	if !market.ID.Equal(id) {
		panic(fmt.Sprintf("marketID at idx %s has wrong ID: %s", id.String(), market.ID.String()))
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

	params := k.GetParams(ctx)
	for _, m := range params.Markets {
		if m.BaseAssetDenom == baseAsset && m.QuoteAssetDenom == quoteAsset {
			return types.Market{}, sdkErrors.Wrap(types.ErrMarketExists, m.String())
		}
	}

	if !k.ccsStorage.HasCurrency(ctx, baseAsset) {
		return types.Market{}, sdkErrors.Wrap(types.ErrWrongAssetDenom, "BaseAsset not registered")
	}
	if !k.ccsStorage.HasCurrency(ctx, quoteAsset) {
		return types.Market{}, sdkErrors.Wrap(types.ErrWrongAssetDenom, "QuoteAsset not registered")
	}

	market := types.NewMarket(k.nextID(params), baseAsset, quoteAsset)
	params.Markets = append(params.Markets, market)
	k.SetParams(ctx, params)

	ctx.EventManager().EmitEvent(types.NewMarketCreatedEvent(market))

	return market, nil
}

// GetList returns all market objects.
func (k Keeper) GetList(ctx sdk.Context) types.Markets {
	k.modulePerms.AutoCheck(types.PermRead)

	return k.GetParams(ctx).Markets
}

// GetListFiltered returns market objects filtered by params.
func (k Keeper) GetListFiltered(ctx sdk.Context, params types.MarketsReq) types.Markets {
	k.modulePerms.AutoCheck(types.PermRead)

	markets := k.GetList(ctx)
	filteredMarkets := make(types.Markets, 0, len(markets))

	for _, m := range markets {
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
	}

	start, end, err := helpers.PaginateSlice(len(filteredMarkets), params.Page, params.Limit)
	if err != nil {
		return types.Markets{}
	}

	return filteredMarkets[start:end]
}
