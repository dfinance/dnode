// Implements multisignature message handler for currency module.
package currencies

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dfinance/dnode/x/core"
	"github.com/dfinance/dnode/x/currencies/msgs"
)

// Handler for currencies messages, proves issue/destroy messages.
func NewMsHandler(keeper Keeper) core.MsHandler {
	return func(ctx sdk.Context, msg core.MsMsg) sdk.Error {
		switch msg := msg.(type) {
		case msgs.MsgIssueCurrency:
			return handleMsMsgIssueCurrency(ctx, keeper, msg)

		default:
			errMsg := fmt.Sprintf("Unrecognized nameservice Msg type: %v", msg.Type())
			return sdk.ErrUnknownRequest(errMsg)
		}
	}
}

// Handle issue message.
func handleMsMsgIssueCurrency(ctx sdk.Context, keeper Keeper, msg msgs.MsgIssueCurrency) sdk.Error {
	return keeper.IssueCurrency(ctx, msg.Symbol, msg.Amount, msg.Decimals, msg.Recipient, msg.IssueID)
}
