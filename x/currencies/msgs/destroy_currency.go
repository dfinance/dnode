// Destroy currency message implementation.
package msgs

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dfinance/dnode/x/currencies/types"
)

// Message for destroy currency
type MsgDestroyCurrency struct {
	ChainID   string         `json:"chainID"`
	Symbol    string         `json:"symbol"`
	Amount    sdk.Int        `json:"amount"`
	Spender   sdk.AccAddress `json:"spender"`
	Recipient string         `json:"recipient"`
}

// Create new message to destroy currency.
func NewMsgDestroyCurrency(chainID, symbol string, amount sdk.Int, spender sdk.AccAddress, recipient string) MsgDestroyCurrency {
	return MsgDestroyCurrency{
		ChainID:   chainID,
		Symbol:    symbol,
		Amount:    amount,
		Spender:   spender,
		Recipient: recipient,
	}
}

// Base route for currencies package.
func (msg MsgDestroyCurrency) Route() string {
	return types.RouterKey
}

// Indeed type to destroy currency.
func (msg MsgDestroyCurrency) Type() string {
	return "destroy_currency"
}

// Validate basic in case of destroy message.
func (msg MsgDestroyCurrency) ValidateBasic() sdk.Error {
	if msg.Spender.Empty() {
		return sdk.ErrInvalidAddress(msg.Spender.String())
	}

	if len(msg.Recipient) == 0 {
		return types.ErrWrongRecipient()
	}

	if len(msg.Symbol) == 0 {
		return types.ErrWrongSymbol(msg.Symbol)
	}

	if msg.Amount.IsZero() {
		return types.ErrWrongAmount(msg.Amount.String())
	}

	// check denom, etc.
	sdk.NewCoin(msg.Symbol, msg.Amount)

	return nil
}

// Get message bytes to sign.
func (msg MsgDestroyCurrency) GetSignBytes() []byte {
	b, err := json.Marshal(msg)

	if err != nil {
		panic(err)
	}

	return sdk.MustSortJSON(b)
}

// Get signers for message.
func (msg MsgDestroyCurrency) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Spender}
}
