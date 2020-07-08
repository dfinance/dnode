package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/dfinance/dnode/x/poa/internal/types"
)

// NewQuerier return keeper querier.
func NewQuerier(poaKeeper Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, error) {
		switch path[0] {
		case types.QueryValidators:
			return queryGetValidators(poaKeeper, ctx)
		case types.QueryValidator:
			return queryGetValidator(poaKeeper, ctx, req)
		case types.QueryMinMax:
			return queryGetMinMax(poaKeeper, ctx)
		default:
			return nil, sdkErrors.Wrapf(sdkErrors.ErrUnknownRequest, "unsupported query endpoint %q for module %q", path[0], types.ModuleName)
		}
	}
}

// queryGetValidators handles getValidators query which returns validator objects.
func queryGetValidators(k Keeper, ctx sdk.Context) ([]byte, error) {
	resp := types.ValidatorsConfirmationsResp{
		Validators:    k.GetValidators(ctx),
		Confirmations: k.GetEnoughConfirmations(ctx),
	}

	bz, err := codec.MarshalJSONIndent(k.cdc, resp)
	if err != nil {
		return nil, sdkErrors.Wrapf(types.ErrInternal, "validatorsConfirmationsRespcould marshal : %v", err)
	}

	return bz, nil
}

// queryGetValidator handles getValidator query which returns validator object.
func queryGetValidator(k Keeper, ctx sdk.Context, req abci.RequestQuery) ([]byte, error) {
	var params types.ValidatorReq
	if err := types.ModuleCdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdkErrors.Wrapf(types.ErrInternal, "failed to parse params: %v", err)
	}

	validator, err := k.GetValidator(ctx, params.Address)
	if err != nil {
		return nil, err
	}

	bz, err := codec.MarshalJSONIndent(k.cdc, validator)
	if err != nil {
		return nil, sdkErrors.Wrapf(types.ErrInternal, "validator marshal: %v", err)
	}

	return bz, nil
}

// queryGetMinMax handles getMinMax query which returns min/max validators amount values.
func queryGetMinMax(k Keeper, ctx sdk.Context) ([]byte, error) {
	bz, err := codec.MarshalJSONIndent(k.cdc, k.GetParams(ctx))
	if err != nil {
		return nil, sdkErrors.Wrapf(types.ErrInternal, "params marshal: %v", err)
	}

	return bz, nil
}
