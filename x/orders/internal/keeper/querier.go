package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/dfinance/dnode/x/orders/internal/types"
)

// NewQuerier return keeper querier.
func NewQuerier(k Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err error) {
		switch path[0] {
		case types.QueryList:
			return queryList(ctx, k, req)
		case types.QueryOrder:
			return queryOrder(ctx, k, req)
		default:
			return nil, sdkErrors.Wrap(sdkErrors.ErrUnknownRequest, "unknown orders query endpoint")
		}
	}
}

// queryList handles list query which return all active order objects filtered.
func queryList(ctx sdk.Context, k Keeper, req abci.RequestQuery) ([]byte, error) {
	var params types.OrdersReq
	if err := k.cdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdkErrors.Wrapf(types.ErrInternal, "failed to parse params: %v", err)
	}

	orders, err := k.GetListFiltered(ctx, params)
	if err != nil {
		return nil, err
	}

	res, err := codec.MarshalJSONIndent(k.cdc, orders)
	if err != nil {
		return nil, fmt.Errorf("orders marshal: %w", err)
	}

	//k.GetLogger(ctx).Debug(fmt.Sprintf("Orders table:\n%s", orders.String()))

	return res, nil
}

// queryOrder handles order query which return order by id.
func queryOrder(ctx sdk.Context, k Keeper, req abci.RequestQuery) ([]byte, error) {
	var params types.OrderReq

	if err := k.cdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdkErrors.Wrapf(types.ErrInternal, "failed to parse params: %v", err)
	}

	order, err := k.Get(ctx, params.ID)
	if err != nil {
		return nil, err
	}

	res, err := codec.MarshalJSONIndent(k.cdc, order)
	if err != nil {
		return nil, fmt.Errorf("order marshal: %w", err)
	}

	return res, nil
}
