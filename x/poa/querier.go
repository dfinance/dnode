// Querier for PoA module.
package poa

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/WingsDao/wings-blockchain/x/poa/types"
)

// Supported queries.
const (
	QueryGetValidators = "validators"
	QueryGetMinMax     = "minmax"
	QueryGetValidator  = "validator"
)

// Creating new Querier.
func NewQuerier(poaKeeper Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, sdk.Error) {
		switch path[0] {
		case QueryGetValidators:
			return queryGetValidators(poaKeeper, ctx)

		case QueryGetMinMax:
			return queryGetMinMax(poaKeeper, ctx)

		case QueryGetValidator:
			return queryGetValidator(poaKeeper, ctx, req)

		default:
			return nil, sdk.ErrUnknownRequest("unknown query")
		}
	}
}

// Query handler for get validators list.
func queryGetValidators(poaKeeper Keeper, ctx sdk.Context) ([]byte, sdk.Error) {
	validators := types.ValidatorsConfirmations{
		Validators:    poaKeeper.GetValidators(ctx),
		Confirmations: poaKeeper.GetEnoughConfirmations(ctx),
	}

	bz, err := codec.MarshalJSONIndent(poaKeeper.cdc, validators)

	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}

	return bz, nil
}

// Query handler for get min/max validators amount values.
func queryGetMinMax(poaKeeper Keeper, ctx sdk.Context) ([]byte, sdk.Error) {
	bz, err := codec.MarshalJSONIndent(poaKeeper.cdc, poaKeeper.GetParams(ctx))

	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}

	return bz, nil
}

// Query handler for get validator by address.
func queryGetValidator(poaKeeper Keeper, ctx sdk.Context, req abci.RequestQuery) ([]byte, sdk.Error) {
	var params types.QueryValidator

	if err := ModuleCdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdk.ErrInternal(fmt.Sprintf("failed to parse params: %s", err))
	}

	validator := poaKeeper.GetValidator(ctx, params.Address)

	bz, err := codec.MarshalJSONIndent(poaKeeper.cdc, validator)
	if err != nil {
		return nil, sdk.ErrInternal(sdk.AppendMsgToErr("could not marshal result to JSON", err.Error()))
	}

	return bz, nil
}
