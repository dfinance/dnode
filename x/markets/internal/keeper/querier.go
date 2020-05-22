package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/dfinance/dnode/x/markets/internal/types"
)

const (
	QueryList   = "list"
	QueryMarket = "market"
)

// NewQuerier return keeper querier.
func NewQuerier(k Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err error) {
		switch path[0] {
		case QueryList:
			return queryList(ctx, k, req)
		case QueryMarket:
			return queryMarket(ctx, k, req)
		default:
			return nil, sdkErrors.Wrap(sdkErrors.ErrUnknownRequest, "unknown markets query endpoint")
		}
	}
}

// queryList handles list query which return all market objects filtered.
func queryList(ctx sdk.Context, k Keeper, req abci.RequestQuery) ([]byte, error) {
	var params types.MarketsReq
	if err := k.cdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdkErrors.Wrapf(types.ErrInternal, "failed to parse params: %v", err)
	}

	markets := k.GetListFiltered(ctx, params)

	res, err := codec.MarshalJSONIndent(k.cdc, markets)
	if err != nil {
		return nil, fmt.Errorf("markets marshal: %w", err)
	}

	//k.GetLogger(ctx).Debug(fmt.Sprintf("Markets table:\n%s", markets.String()))

	return res, nil
}

// queryMarket handles order query which return market by id.
func queryMarket(ctx sdk.Context, k Keeper, req abci.RequestQuery) ([]byte, error) {
	var params types.MarketReq

	if err := k.cdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdkErrors.Wrapf(types.ErrInternal, "failed to parse params: %v", err)
	}

	order, err := k.Get(ctx, params.ID)
	if err != nil {
		return nil, err
	}

	res, err := codec.MarshalJSONIndent(k.cdc, order)
	if err != nil {
		return nil, fmt.Errorf("market marshal: %w", err)
	}

	return res, nil
}
