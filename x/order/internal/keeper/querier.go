package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	abci "github.com/tendermint/tendermint/abci/types"
)

const (
	QueryList = "list"
)

// NewQuerier return keeper querier.
func NewQuerier(k Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err error) {
		switch path[0] {
		case QueryList:
			return queryList(ctx, k)
		default:
			return nil, sdkErrors.Wrap(sdkErrors.ErrUnknownRequest, "unknown order query endpoint")
		}
	}
}

// queryList handles list query which return all active order objects.
func queryList(ctx sdk.Context, k Keeper) ([]byte, error) {
	orders, err := k.List(ctx)
	if err != nil {
		return nil, err
	}

	res, err := codec.MarshalJSONIndent(k.cdc, orders)
	if err != nil {
		return nil, fmt.Errorf("orders marshal: %w", err)
	}

	k.GetLogger(ctx).Debug(fmt.Sprintf("Orders table:\n%s", orders.String()))

	return res, nil
}
