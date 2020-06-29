package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/dfinance/dnode/x/currencies/internal/types"
)

// NewQuerier return keeper querier.
func NewQuerier(k Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		switch path[0] {
		case types.QueryDestroys:
			return queryGetDestroys(k, ctx, req)
		case types.QueryDestroy:
			return queryGetDestroy(k, ctx, req)
		case types.QueryIssue:
			return queryGetIssue(k, ctx, req)
		case types.QueryCurrency:
			return queryGetCurrency(k, ctx, req)
		default:
			return nil, sdkErrors.Wrap(sdkErrors.ErrUnknownRequest, "unknown currencies query endpoint")
		}
	}
}

// queryGetDestroys handles getDestroys query which return destroy objects filtered.
func queryGetDestroys(k Keeper, ctx sdk.Context, req abci.RequestQuery) ([]byte, error) {
	params := types.DestroysReq{}
	if err := types.ModuleCdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdkErrors.Wrapf(types.ErrInternal, "failed to parse params: %v", err)
	}

	destroys, err := k.GetDestroysFiltered(ctx, params)
	if err != nil {
		return nil, sdkErrors.Wrap(types.ErrInternal, err.Error())
	}

	bz, err := codec.MarshalJSONIndent(k.cdc, destroys)
	if err != nil {
		return nil, sdkErrors.Wrapf(types.ErrInternal, "destroys marshal: %v", err)
	}

	return bz, nil
}

// queryGetDestroy handles getDestroy query which return destroy by id.
func queryGetDestroy(k Keeper, ctx sdk.Context, req abci.RequestQuery) ([]byte, error) {
	params := types.DestroyReq{}
	if err := types.ModuleCdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdkErrors.Wrapf(types.ErrInternal, "failed to parse params: %v", err)
	}

	destroy, err := k.GetDestroy(ctx, params.ID)
	if err != nil {
		return nil, err
	}

	bz, err := codec.MarshalJSONIndent(k.cdc, destroy)
	if err != nil {
		return nil, sdkErrors.Wrapf(types.ErrInternal, "destroy marshal: %v", err)
	}

	return bz, nil
}

// queryGetIssue handles getIssue query which return issue by id.
func queryGetIssue(k Keeper, ctx sdk.Context, req abci.RequestQuery) ([]byte, error) {
	params := types.IssueReq{}
	if err := types.ModuleCdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdkErrors.Wrapf(types.ErrInternal, "failed to parse params: %v", err)
	}

	issue, err := k.GetIssue(ctx, params.ID)
	if err != nil {
		return nil, err
	}

	bz, err := codec.MarshalJSONIndent(k.cdc, issue)
	if err != nil {
		return nil, sdkErrors.Wrapf(types.ErrInternal, "issue marshal: %v", err)
	}

	return bz, nil
}

// queryGetCurrency handles getCurrency query which return currency by denom.
// Query handler to get currency by symbol.
func queryGetCurrency(k Keeper, ctx sdk.Context, req abci.RequestQuery) ([]byte, error) {
	params := types.CurrencyReq{}
	if err := types.ModuleCdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdkErrors.Wrapf(types.ErrInternal, "failed to parse params: %v", err)
	}

	currency, err := k.GetCurrency(ctx, params.Denom)
	if err != nil {
		return nil, err
	}

	bz, err := codec.MarshalJSONIndent(k.cdc, currency)
	if err != nil {
		return nil, sdkErrors.Wrapf(types.ErrInternal, "currency marshal: %v", err)
	}

	return bz, nil
}
