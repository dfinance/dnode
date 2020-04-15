// Issue currency message implementation.
package msgs

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/dfinance/dnode/x/currencies/types"
)

// Msg struct for issue new currencies.
// IssueID could be txHash of transaction in another blockchain.
type MsgIssueCurrency struct {
	Symbol    string         `json:"symbol"`
	Amount    sdk.Int        `json:"amount"`
	Decimals  int8           `json:"decimals"`
	Recipient sdk.AccAddress `json:"recipient"`
	IssueID   string         `json:"issueID"`
}

// Create new issue currency message.
func NewMsgIssueCurrency(symbol string, amount sdk.Int, decimals int8, recipient sdk.AccAddress, issueID string) MsgIssueCurrency {
	return MsgIssueCurrency{
		Symbol:    symbol,
		Amount:    amount,
		Decimals:  decimals,
		Recipient: recipient,
		IssueID:   issueID,
	}
}

// Common router for currencies package.
func (msg MsgIssueCurrency) Route() string {
	return types.RouterKey
}

// Command for issue new currencies.
func (msg MsgIssueCurrency) Type() string {
	return "issue_currency"
}

// Basic validation, without state.
func (msg MsgIssueCurrency) ValidateBasic() error {
	if msg.Recipient.Empty() {
		return sdkErrors.Wrap(sdkErrors.ErrInvalidAddress, "empty recipient")
	}

	if len(msg.Symbol) == 0 {
		return sdkErrors.Wrap(types.ErrWrongSymbol, "empty")
	}

	if msg.Decimals < 0 {
		return types.ErrWrongDecimals
	}

	if msg.Amount.IsZero() {
		return types.ErrWrongAmount
	}

	if len(msg.IssueID) == 0 {
		return sdkErrors.Wrap(types.ErrWrongIssueID, "empty")
	}

	// lets try to create coin and validate denom
	sdk.NewCoin(msg.Symbol, msg.Amount)

	return nil
}

// Getting bytes for signature.
func (msg MsgIssueCurrency) GetSignBytes() []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}

	return sdk.MustSortJSON(b)
}

// Check who should sign message.
func (msg MsgIssueCurrency) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{}
}
