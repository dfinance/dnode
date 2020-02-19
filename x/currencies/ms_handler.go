// Implements multisignature message handler for currency module.
package currencies

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/WingsDao/wings-blockchain/x/core"
	"github.com/WingsDao/wings-blockchain/x/currencies/msgs"
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
	err := keeper.IssueCurrency(ctx, msg.Symbol, msg.Amount, msg.Decimals, msg.Recipient, msg.IssueID)
	return err
}
