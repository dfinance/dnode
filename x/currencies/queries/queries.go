package queries

import (
	"wings-blockchain/x/currencies"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"wings-blockchain/x/currencies/types"
)

const (
	QueryGetCurrency  = "currency"
	QueryGetIssue  	  = "issue"
	QueryGetDestroys  = "destroys"
	QueryGetDestroy   = "destroy"
)

// Querier for currencies module
func NewQuerier(keeper currencies.Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, sdk.Error) {
		switch path[0] {
		case QueryGetIssue:
			return queryGetIssue(keeper, ctx, path[1:])

		case QueryGetCurrency:
			return queryGetCurrency(keeper, ctx, path[1:])

        case QueryGetDestroys:
            return queryGetDestroys(keeper, ctx, path[1:])

        case QueryGetDestroy:
            return queryGetDestroy(keeper, ctx, path[1:])

		default:
			return nil, sdk.ErrUnknownRequest("unknown query")
		}
	}
}

// Query handler to get currency
func queryGetCurrency(keeper currencies.Keeper, ctx sdk.Context, params []string) ([]byte, sdk.Error) {
	getCurRes := QueryCurrencyRes{}
	symbol   := params[0]

	cur := keeper.GetCurrency(ctx, symbol)

	if cur.Symbol != symbol {
		return []byte{}, types.ErrNotExistCurrency(symbol)
	}

	getCurRes.Currency = cur

	bz, err := codec.MarshalJSONIndent(keeper.GetCDC(), getCurRes)

	if err != nil {
		panic(err)
	}

	return bz, nil
}

// Query handler to get currencies
func queryGetIssue(keeper currencies.Keeper, ctx sdk.Context, params []string) ([]byte, sdk.Error) {
    issueID  := params[0]

	issueRes := QueryIssueRes{}
	issue    := keeper.GetIssue(ctx, issueID)

	if issue.Recipient.Empty() {
	    return []byte{}, types.ErrWrongIssueID(issueID)
    }

	issueRes.Issue = issue

	bz, err := codec.MarshalJSONIndent(keeper.GetCDC(), issueRes)

	if err != nil {
		panic(err)
	}

	return bz, nil
}

// Get destroy from API
func queryGetDestroy(keeper currencies.Keeper, ctx sdk.Context, params []string) ([]byte, sdk.Error) {
    var destroyRes QueryDestroyRes

    id, _   := sdk.NewIntFromString(params[0])

    destroyRes.Destroy = keeper.GetDestroy(ctx, id)

    bz, err := codec.MarshalJSONIndent(keeper.GetCDC(), destroyRes)

    if err != nil {
        panic(err)
    }

    return bz, nil
}

// Get destroys from API
func queryGetDestroys(keeper currencies.Keeper, ctx sdk.Context, params []string) ([]byte, sdk.Error) {
    destroys := make(QueryDestroysRes, 0)

    page,  _  := sdk.NewIntFromString(params[0])
    limit, _  := sdk.NewIntFromString(params[1])

    start := page.SubRaw(1).Mul(limit)
    end   := start.Add(limit)

    for ; start.LT(end) && keeper.HasDestroy(ctx, start); start = start.AddRaw(1) {
        destroy  := keeper.GetDestroy(ctx, start)
        destroys = append(destroys, QueryDestroyRes{Destroy: destroy})
    }

    bz, err := codec.MarshalJSONIndent(keeper.GetCDC(), destroys)

    if err != nil {
        panic(err)
    }

    return bz, nil
}
