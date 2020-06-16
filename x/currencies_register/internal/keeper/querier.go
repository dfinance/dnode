package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/dfinance/dnode/x/currencies_register/internal/types"
)

const (
	QueryInfo  = "info"
)

// NewQuerier return keeper querier.
func NewQuerier(k Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err error) {
		switch path[0] {
		case QueryInfo:
			return queryInfo(ctx, k, req)
		default:
			return nil, sdkErrors.Wrap(sdkErrors.ErrUnknownRequest, "unknown currencies_register query endpoint")
		}
	}
}

// queryInfo handles CurrencyInfo query.
func queryInfo(ctx sdk.Context, k Keeper, req abci.RequestQuery) ([]byte, error) {
	var params types.CurrencyInfoReq

	if err := k.cdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdkErrors.Wrapf(types.ErrInternal, "failed to parse params: %v", err)
	}

	ccInfo, err := k.GetCurrencyInfo(ctx, params.Denom)
	if err != nil {
		return nil, err
	}

	res, err := codec.MarshalJSONIndent(k.cdc, ccInfo)
	if err != nil {
		return nil, fmt.Errorf("ccInfo marshal: %w", err)
	}

	return res, nil
}
