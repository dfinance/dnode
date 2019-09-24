package currencies


import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"fmt"
	"wings-blockchain/x/currencies/msgs"
	msTypes "wings-blockchain/x/multisig/types"
)

// Handler for currencies messages, provess issue/destroy messages
func NewMsHandler(keeper Keeper) msTypes.MsHandler {
	return func (ctx sdk.Context, msg msTypes.MsMsg) sdk.Error {
		switch msg := msg.(type) {
		case msgs.MsgIssueCurrency:
			return handleMsMsgIssueCurrency(ctx, keeper, msg)

		default:
			errMsg := fmt.Sprintf("Unrecognized nameservice Msg type: %v", msg.Type())
			return sdk.ErrUnknownRequest(errMsg)
		}
	}
}

// Handle issue message
func handleMsMsgIssueCurrency(ctx sdk.Context, keeper Keeper, msg msgs.MsgIssueCurrency) sdk.Error {
	err := keeper.IssueCurrency(ctx, msg.CurrencyId, msg.Symbol, msg.Amount, msg.Decimals, msg.Recipient, msg.IssueID)
	return err
}
