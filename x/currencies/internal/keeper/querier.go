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
		case types.QueryWithdraws:
			return queryGetWithdraws(k, ctx, req)
		case types.QueryWithdraw:
			return queryGetWithdraw(k, ctx, req)
		case types.QueryIssue:
			return queryGetIssue(k, ctx, req)
		case types.QueryCurrency:
			return queryGetCurrency(k, ctx, req)
		default:
			return nil, sdkErrors.Wrapf(sdkErrors.ErrUnknownRequest, "unsupported query endpoint %q for module %q", path[0], types.ModuleName)
		}
	}
}

// queryGetWithdraws handles getWithdraws query which return withdraw objects filtered.
func queryGetWithdraws(k Keeper, ctx sdk.Context, req abci.RequestQuery) ([]byte, error) {
	params := types.WithdrawsReq{}
	if err := types.ModuleCdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdkErrors.Wrapf(types.ErrInternal, "failed to parse params: %v", err)
	}

	withdraws, err := k.GetWithdrawsFiltered(ctx, params)
	if err != nil {
		return nil, sdkErrors.Wrap(types.ErrInternal, err.Error())
	}

	bz, err := codec.MarshalJSONIndent(k.cdc, withdraws)
	if err != nil {
		return nil, sdkErrors.Wrapf(types.ErrInternal, "withdraws marshal: %v", err)
	}

	return bz, nil
}

// queryGetWithdraw handles getWithdraw query which return withdraw by id.
func queryGetWithdraw(k Keeper, ctx sdk.Context, req abci.RequestQuery) ([]byte, error) {
	params := types.WithdrawReq{}
	if err := types.ModuleCdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdkErrors.Wrapf(types.ErrInternal, "failed to parse params: %v", err)
	}

	withdraw, err := k.GetWithdraw(ctx, params.ID)
	if err != nil {
		return nil, err
	}

	bz, err := codec.MarshalJSONIndent(k.cdc, withdraw)
	if err != nil {
		return nil, sdkErrors.Wrapf(types.ErrInternal, "withdraw marshal: %v", err)
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

	currency, err := k.ccsKeeper.GetCurrency(ctx, params.Denom)
	if err != nil {
		return nil, err
	}

	bz, err := codec.MarshalJSONIndent(k.cdc, currency)
	if err != nil {
		return nil, sdkErrors.Wrapf(types.ErrInternal, "currency marshal: %v", err)
	}

	return bz, nil
}
