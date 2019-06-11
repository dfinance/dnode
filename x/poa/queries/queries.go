package queries

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"wings-blockchain/x/poa"
	"github.com/cosmos/cosmos-sdk/codec"
)

const (
	QueryGetValidators = "validators"
	QueryGetMinMax     = "minmax"
	QueryGetValidator  = "validator"
)

// Querier for PoA module
func NewQuerier(keeper poa.Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, sdk.Error) {
		switch path[0] {
		case QueryGetValidators:
			return queryGetValidators(keeper, ctx)

		case QueryGetMinMax:
			return queryGetMinMax(keeper, ctx)

		case QueryGetValidator:
			return queryGetValidator(keeper, ctx, path[1:])

		default:
			return nil, sdk.ErrUnknownRequest("unknown query")
		}
	}
}

// Query handler for get validators list
func queryGetValidators(keeper poa.Keeper, ctx sdk.Context) ([]byte, sdk.Error) {
	validatorsRes := QueryValidatorsRes{}

	validatorsRes.Validators 	= keeper.GetValidators(ctx)
	validatorsRes.Amount 		= len(validatorsRes.Validators)
	validatorsRes.Confirmations = int(keeper.GetEnoughConfirmations(ctx))

	bz, err := codec.MarshalJSONIndent(keeper.GetCDC(), validatorsRes)

	if err != nil {
		panic(err)
	}

	return bz, nil
}

// Query handler for get min/max validators amount values
func queryGetMinMax(keeper poa.Keeper, ctx sdk.Context) ([]byte, sdk.Error) {
	minMaxRes := QueryMinMaxRes{}

	minMaxRes.Min = int(keeper.GetMinValidators(ctx))
	minMaxRes.Max = int(keeper.GetMaxValidators(ctx))

	bz, err := codec.MarshalJSONIndent(keeper.GetCDC(), minMaxRes)

	if err != nil {
		panic(err)
	}

	return bz, nil
}

// Query handler for get validator by address
func queryGetValidator(keeper poa.Keeper, ctx sdk.Context, params []string) ([]byte, sdk.Error) {
	getValidatorRes := QueryGetValidatorRes{}

	acc, err := sdk.AccAddressFromBech32(params[1])

	if err != nil {
		return nil, sdk.ErrInvalidAddress(params[1])
	}

	getValidatorRes.validator = keeper.GetValidator(ctx, acc)

	bz, err := codec.MarshalJSONIndent(keeper.GetCDC(), getValidatorRes)

	if err != nil {
		panic(err)
	}

	return bz, nil
}