package vmauth

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	authTypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// NewQuerier creates a querier for auth REST endpoints.
func NewQuerier(keeper VMAccountKeeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		switch path[0] {
		case authTypes.QueryAccount:
			return queryAccount(ctx, req, keeper)
		default:
			return nil, sdkErrors.Wrap(sdkErrors.ErrUnknownRequest, "unknown auth query endpoint")
		}
	}
}

// queryAccount is an account getter querier handler.
func queryAccount(ctx sdk.Context, req abci.RequestQuery, keeper VMAccountKeeper) ([]byte, error) {
	var params authTypes.QueryAccountParams
	if err := keeper.cdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdkErrors.Wrapf(ErrInternal, "failed to parse params: %v", err)
	}

	account := keeper.GetAccount(ctx, params.Address)
	if account == nil {
		return nil, sdkErrors.Wrapf(sdkErrors.ErrUnknownAddress, "account %q does not exist", params.Address)
	}

	bz, err := codec.MarshalJSONIndent(keeper.cdc, account)
	if err != nil {
		return nil, sdkErrors.Wrapf(ErrInternal, "could not marshal result to JSON: %v", err)
	}

	return bz, nil
}
