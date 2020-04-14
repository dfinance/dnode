// Implements multisignature message handler for currency module.
package currencies

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/dfinance/dnode/x/core"
	"github.com/dfinance/dnode/x/currencies/msgs"
)

// Handler for currencies messages, proves issue/destroy messages.
func NewMsHandler(keeper Keeper) core.MsHandler {
	return func(ctx sdk.Context, msg core.MsMsg) error {
		switch msg := msg.(type) {
		case msgs.MsgIssueCurrency:
			return handleMsMsgIssueCurrency(ctx, keeper, msg)

		default:
			return sdkErrors.Wrapf(sdkErrors.ErrUnknownRequest, "unrecognized nameservice Msg type: %v", msg.Type())
		}
	}
}

// Handle issue message.
func handleMsMsgIssueCurrency(ctx sdk.Context, keeper Keeper, msg msgs.MsgIssueCurrency) error {
	return keeper.IssueCurrency(ctx, msg.Symbol, msg.Amount, msg.Decimals, msg.Recipient, msg.IssueID)
}
