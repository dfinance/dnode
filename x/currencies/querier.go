// Implements querier for currency module.
package currencies

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/dfinance/dnode/x/currencies/types"
)

const (
	QueryGetDestroys = "destroys"
	QueryGetDestroy  = "destroy"
	QueryGetIssue    = "issue"
	QueryGetCurrency = "currency"
)

// Creating new querier.
func NewQuerier(ccKeeper Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, sdk.Error) {
		switch path[0] {
		case QueryGetDestroys:
			return queryGetDestroys(ccKeeper, ctx, req)

		case QueryGetDestroy:
			return queryGetDestroy(ccKeeper, ctx, req)

		case QueryGetIssue:
			return queryGetIssue(ccKeeper, ctx, req)

		case QueryGetCurrency:
			return queryGetCurrency(ccKeeper, ctx, req)

		default:
			return nil, sdk.ErrUnknownRequest("unknown query")
		}
	}
}

// Query handler to get destroys by id.
func queryGetDestroys(ccKeeper Keeper, ctx sdk.Context, req abci.RequestQuery) ([]byte, sdk.Error) {
	var params types.DestroysReq

	if err := ModuleCdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}

	start := params.Page.SubRaw(1).Mul(params.Limit)
	end := start.Add(params.Limit)

	destroys := make(types.Destroys, 0)

	for ; start.LT(end) && ccKeeper.HasDestroy(ctx, start); start = start.AddRaw(1) {
		destroy := ccKeeper.GetDestroy(ctx, start)
		destroys = append(destroys, destroy)
	}

	bz, err := codec.MarshalJSONIndent(ccKeeper.cdc, destroys)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}

	return bz, nil
}

// Query handler to get destroy by id.
func queryGetDestroy(ccKeeper Keeper, ctx sdk.Context, req abci.RequestQuery) ([]byte, sdk.Error) {
	var params types.DestroyReq

	if err := ModuleCdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}

	destroy := ccKeeper.GetDestroy(ctx, params.DestroyId)

	bz, err := codec.MarshalJSONIndent(ccKeeper.cdc, destroy)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}

	return bz, nil
}

// Query handler to get issue by id.
func queryGetIssue(ccKeeper Keeper, ctx sdk.Context, req abci.RequestQuery) ([]byte, sdk.Error) {
	var params types.IssueReq

	if err := ModuleCdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}

	issue := ccKeeper.GetIssue(ctx, params.IssueID)
	if issue.Recipient.Empty() {
		return nil, types.ErrWrongIssueID(params.IssueID)
	}

	bz, err := codec.MarshalJSONIndent(ccKeeper.cdc, issue)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}

	return bz, nil
}

// Query handler to get currency by symbol.
func queryGetCurrency(ccKeeper Keeper, ctx sdk.Context, req abci.RequestQuery) ([]byte, sdk.Error) {
	var params types.CurrencyReq

	if err := ModuleCdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}

	currency := ccKeeper.getCurrency(ctx, params.Symbol)
	if currency.Symbol != params.Symbol {
		return []byte{}, types.ErrNotExistCurrency(params.Symbol)
	}

	bz, err := codec.MarshalJSONIndent(ccKeeper.cdc, currency)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}

	return bz, nil
}
