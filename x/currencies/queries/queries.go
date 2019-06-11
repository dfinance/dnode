package queries

import (
	"wings-blockchain/x/currencies"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/cosmos/cosmos-sdk/codec"
)

const (
	QueryGetCurrency   = "currency"
	QueryDenoms  	   = "denoms"
)

// Querier for currencies module
func NewQuerier(keeper currencies.Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, sdk.Error) {
		switch path[0] {
		case QueryDenoms:
			return queryGetDenoms(keeper, ctx)

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

	symbol := params[1]
	cur := keeper.GetCurrency(ctx, symbol)

	getCurRes.Currency = cur

	bz, err := codec.MarshalJSONIndent(keeper.GetCDC(), getCurRes)

	if err != nil {
		panic(err)
	}

	return bz, nil
}

// Query handler to get denoms
func queryGetDenoms(keeper currencies.Keeper, ctx sdk.Context) ([]byte, sdk.Error) {
	denomsRes := QueryDenomsRes{}
	denomsRes.Denoms = keeper.GetDenoms(ctx)

	bz, err := codec.MarshalJSONIndent(keeper.GetCDC(), denomsRes)

	if err != nil {
		panic(err)
	}

	return bz, nil
}
