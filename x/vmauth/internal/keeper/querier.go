package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	authTypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/dfinance/dnode/x/vmauth/internal/types"
)

// NewQuerier return keeper querier.
func NewQuerier(k VMAccountKeeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		switch path[0] {
		case authTypes.QueryAccount:
			return queryAccount(ctx, req, k)
		default:
			return nil, sdkErrors.Wrapf(sdkErrors.ErrUnknownRequest, "unsupported query endpoint %q for module %q", path[0], types.ModuleName)
		}
	}
}

// queryAccount handles account getter query.
func queryAccount(ctx sdk.Context, req abci.RequestQuery, k VMAccountKeeper) ([]byte, error) {
	var params authTypes.QueryAccountParams
	if err := k.cdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdkErrors.Wrapf(types.ErrInternal, "failed to parse params: %v", err)
	}

	account := k.GetAccount(ctx, params.Address)
	if account == nil {
		return nil, sdkErrors.Wrapf(sdkErrors.ErrUnknownAddress, "account %q: not found", params.Address)
	}

	bz, err := codec.MarshalJSONIndent(k.cdc, account)
	if err != nil {
		return nil, sdkErrors.Wrapf(types.ErrInternal, "AccAddress marshal: %v", err)
	}

	return bz, nil
}
