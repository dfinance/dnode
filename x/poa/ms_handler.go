package poa

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/dfinance/dnode/x/core/msmodule"
	"github.com/dfinance/dnode/x/poa/internal/keeper"
)

// NewMsHandler creates core.MsMsg type messages handler.
func NewMsHandler(k keeper.Keeper) msmodule.MsHandler {
	return func(ctx sdk.Context, msg msmodule.MsMsg) error {
		switch msg := msg.(type) {
		case MsgAddValidator:
			return handleMsMsgAddValidator(ctx, k, msg)
		case MsgRemoveValidator:
			return handleMsMsgRemoveValidator(ctx, k, msg)
		case MsgReplaceValidator:
			return handleMsMsgReplaceValidator(ctx, k, msg)
		default:
			return sdkErrors.Wrapf(sdkErrors.ErrUnknownRequest, "unrecognized %s module multisig msg type: %v", ModuleName, msg.Type())
		}
	}
}

// handleMsMsgAddValidator hanldes MsgAddValidator multisig message.
func handleMsMsgAddValidator(ctx sdk.Context, k keeper.Keeper, msg MsgAddValidator) error {
	if err := k.AddValidator(ctx, msg.Address, msg.EthAddress); err != nil {
		return err
	}

	return nil
}

// handleMsMsgRemoveValidator hanldes MsgRemoveValidator multisig message.
// Handle MsgRemoveValidator for remove validator.
func handleMsMsgRemoveValidator(ctx sdk.Context, k keeper.Keeper, msg MsgRemoveValidator) error {
	if err := k.RemoveValidator(ctx, msg.Address); err != nil {
		return err
	}

	return nil
}

// handleMsMsgReplaceValidator hanldes MsgReplaceValidator multisig message.
func handleMsMsgReplaceValidator(ctx sdk.Context, k keeper.Keeper, msg MsgReplaceValidator) error {
	if err := k.ReplaceValidator(ctx, msg.OldValidator, msg.NewValidator, msg.EthAddress); err != nil {
		return err
	}

	return nil
}
