package queries

import (
	"wings-blockchain/x/currencies"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"wings-blockchain/x/currencies/types"
)

const (
	QueryGetCurrency = "get"
	QueryIssue  	 = "issue"
)

// Querier for currencies module
func NewQuerier(keeper currencies.Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, sdk.Error) {
		switch path[0] {
		case QueryIssue:
			return queryGetIssue(keeper, ctx, path[1:])

		case QueryGetCurrency:
			return queryGetCurrency(keeper, ctx, path[1:])

		default:
			return nil, sdk.ErrUnknownRequest("unknown query")
		}
	}
}

// Query handler to get currency
func queryGetCurrency(keeper currencies.Keeper, ctx sdk.Context, params []string) ([]byte, sdk.Error) {
	getCurRes := QueryCurrencyRes{}

	symbol := params[0]

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
