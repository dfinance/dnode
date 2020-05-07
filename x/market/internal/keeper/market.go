package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/market/internal/types"
)

func (k Keeper) Has(ctx sdk.Context, id dnTypes.ID) bool {
	_, err := k.Get(ctx, id)

	return err == nil
}

func (k Keeper) Get(ctx sdk.Context, id dnTypes.ID) (types.Market, error) {
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

func (k Keeper) Add(ctx sdk.Context, nominee, baseAsset, quoteAsset string, baseDecimals uint8) (types.Market, error) {
	if !k.isNominee(ctx, nominee) {
		return types.Market{}, sdkErrors.Wrap(types.ErrWrongNominee, nominee)
	}

	params := k.GetParams(ctx)
	for _, m := range params.Markets {
		if m.BaseAssetDenom == baseAsset && m.QuoteAssetDenom == quoteAsset {
			return types.Market{}, sdkErrors.Wrap(types.ErrMarketExists, m.String())
		}
	}

	market := types.NewMarket(k.nextID(params), baseAsset, quoteAsset, baseDecimals)
	params.Markets = append(params.Markets, market)
	k.SetParams(ctx, params)

	return market, nil
}
