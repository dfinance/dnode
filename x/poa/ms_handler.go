package poa

import (
	"wings-blockchain/x/poa/types"
	"wings-blockchain/x/poa/msgs"
	ms "wings-blockchain/x/multisig/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"fmt"
)

// New message handler for PoA module
func NewMsHandler(keeper Keeper) ms.MsHandler {
	return func (ctx sdk.Context, msg ms.MsMsg) sdk.Error {
		switch msg := msg.(type) {
		case msgs.MsgAddValidator:
			return handleMsMsgAddValidator(ctx, keeper, msg)

		case msgs.MsgReplaceValidator:
			return handleMsMsgReplaceValidator(ctx, keeper, msg)

		case msgs.MsgRemoveValidator:
			return handleMsMsgRemoveValidator(ctx, keeper, msg)

		default:
			errMsg := fmt.Sprintf("Unrecognized nameservice Msg type: %v", msg.Type())
			return sdk.ErrUnknownRequest(errMsg)
		}
	}
}

// Handle MsgAddValidator for add new validator
func handleMsMsgAddValidator(ctx sdk.Context, keeper Keeper, msg msgs.MsgAddValidator) sdk.Error {
	if keeper.HasValidator(ctx, msg.Address) {
		return types.ErrValidatorExists(msg.Address.String())
	}

	maxValidators := keeper.GetMaxValidators(ctx)
	amount        := keeper.GetValidatorAmount(ctx)

	if amount + 1 > maxValidators {
		return types.ErrMaxValidatorsReached(maxValidators)
	}

	keeper.AddValidator(ctx, msg.Address, msg.EthAddress)
	return nil
}

// Handle MsgRemoveValidator for remove validator
func handleMsMsgRemoveValidator(ctx sdk.Context, keeper Keeper, msg msgs.MsgRemoveValidator) sdk.Error {
	if !keeper.HasValidator(ctx, msg.Address) {
		return types.ErrValidatorDoesntExists(msg.Address.String())
	}

	minValidators := keeper.GetMinValidators(ctx)
	amount		  := keeper.GetValidatorAmount(ctx)

	if amount - 1 < minValidators {
		return types.ErrMinValidatorsReached(minValidators)
	}

	keeper.RemoveValidator(ctx, msg.Address)
	return nil
}

// Handle MsgReplaceValidator for replace validator
func handleMsMsgReplaceValidator(ctx sdk.Context, keeper Keeper, msg msgs.MsgReplaceValidator) sdk.Error {
	if !keeper.HasValidator(ctx, msg.OldValidator) {
		return types.ErrValidatorDoesntExists(msg.OldValidator.String())
	}

	if keeper.HasValidator(ctx, msg.NewValidator) {
		return types.ErrValidatorExists(msg.NewValidator.String())
	}

	keeper.ReplaceValidator(ctx, msg.OldValidator, msg.NewValidator, msg.EthAddress)
	return nil
}