package poa

import (
	"wings-blockchain/x/poa/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (keeper Keeper) InitGenesis(ctx sdk.Context, validators []*types.Validator) sdk.Error {
	minValidators := keeper.GetMinValidators(ctx)

	if len(validators) < int(minValidators) {
		return types.ErrNotEnoungValidators(uint16(len(validators)), minValidators)
	}

	for _, val := range validators {
		keeper.AddValidator(ctx, val.Address, val.EthAddress)
	}

	return nil
}