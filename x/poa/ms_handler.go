// Multisignature handler for processing multisignature messages like: add, remove, replace validator.
package poa

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/dfinance/dnode/x/core/msmodule"
	"github.com/dfinance/dnode/x/poa/msgs"
	"github.com/dfinance/dnode/x/poa/types"
)

// New multisignature message handler for PoA module.
func NewMsHandler(keeper Keeper) msmodule.MsHandler {
	return func(ctx sdk.Context, msg msmodule.MsMsg) error {
		switch msg := msg.(type) {
		case msgs.MsgAddValidator:
			return handleMsMsgAddValidator(ctx, keeper, msg)

		case msgs.MsgReplaceValidator:
			return handleMsMsgReplaceValidator(ctx, keeper, msg)

		case msgs.MsgRemoveValidator:
			return handleMsMsgRemoveValidator(ctx, keeper, msg)

		default:
			return sdkErrors.Wrapf(sdkErrors.ErrUnknownRequest, "unrecognized nameservice Msg type: %v", msg.Type())
		}
	}
}

// Handle MsgAddValidator for add new validator.
func handleMsMsgAddValidator(ctx sdk.Context, keeper Keeper, msg msgs.MsgAddValidator) error {
	if keeper.HasValidator(ctx, msg.Address) {
		return sdkErrors.Wrap(types.ErrValidatorExists, msg.Address.String())
	}

	maxValidators := keeper.GetMaxValidators(ctx)
	amount := keeper.GetValidatorAmount(ctx)

	if amount+1 > maxValidators {
		return sdkErrors.Wrapf(types.ErrMaxValidatorsReached, "%d",maxValidators)
	}

	keeper.AddValidator(ctx, msg.Address, msg.EthAddress)
	return nil
}

// Handle MsgRemoveValidator for remove validator.
func handleMsMsgRemoveValidator(ctx sdk.Context, keeper Keeper, msg msgs.MsgRemoveValidator) error {
	if !keeper.HasValidator(ctx, msg.Address) {
		return sdkErrors.Wrap(types.ErrValidatorDoesntExists, msg.Address.String())
	}

	minValidators := keeper.GetMinValidators(ctx)
	amount := keeper.GetValidatorAmount(ctx)

	if amount-1 < minValidators {
		return sdkErrors.Wrapf(types.ErrMinValidatorsReached, "%d", minValidators)
	}

	keeper.RemoveValidator(ctx, msg.Address)

	return nil
}

// Handle MsgReplaceValidator for replace validator.
func handleMsMsgReplaceValidator(ctx sdk.Context, keeper Keeper, msg msgs.MsgReplaceValidator) error {
	if !keeper.HasValidator(ctx, msg.OldValidator) {
		return sdkErrors.Wrap(types.ErrValidatorDoesntExists, msg.OldValidator.String())
	}

	if keeper.HasValidator(ctx, msg.NewValidator) {
		return sdkErrors.Wrap(types.ErrValidatorExists, msg.NewValidator.String())
	}

	keeper.ReplaceValidator(ctx, msg.OldValidator, msg.NewValidator, msg.EthAddress)

	return nil
}
