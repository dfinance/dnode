package currencies

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/core/msmodule"
	"github.com/dfinance/dnode/x/currencies/internal/keeper"
)

// NewMsHandler creates core.MsMsg type messages handler.
func NewMsHandler(keeper keeper.Keeper) msmodule.MsHandler {
	return func(ctx sdk.Context, msg msmodule.MsMsg) error {
		switch msg := msg.(type) {
		case MsgIssueCurrency:
			return handleMsMsgIssueCurrency(ctx, keeper, msg)

		case MsgUnstakeCurrency:
			return handleMsMsgUnstakeCurrency(ctx, keeper, msg)

		default:
			return sdkErrors.Wrapf(sdkErrors.ErrUnknownRequest, "unrecognized %s module multisig msg type: %v", ModuleName, msg.Type())
		}
	}
}

func handleMsMsgUnstakeCurrency(ctx sdk.Context, keeper keeper.Keeper, msg MsgUnstakeCurrency) error {
	if err := keeper.UnstakeCurrency(ctx, msg.Staker); err != nil {
		return err
	}

	return nil
}

// handleMsMsgIssueCurrency hanldes MsgIssueCurrency multisig message.
func handleMsMsgIssueCurrency(ctx sdk.Context, keeper keeper.Keeper, msg MsgIssueCurrency) error {
	if err := keeper.IssueCurrency(ctx, msg.ID, msg.Coin, msg.Payee); err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(dnTypes.NewModuleNameEvent(ModuleName))

	return nil
}
