package msgs

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"encoding/json"
	"wings-blockchain/x/currencies/types"
)

// Message for destory currency
type MsgDestroyCurrency struct {
	Symbol 	 string			`json:"symbol"`
	Amount   sdk.Int	 	`json:"amount"`
	Sender   sdk.AccAddress `json:"sender"`
}

// Create new message to destory currency
func NewMsgDestroyCurrency(symbol string, amount sdk.Int, sender sdk.AccAddress) MsgDestroyCurrency {
	return MsgDestroyCurrency{
		Symbol: symbol,
		Amount: amount,
		Sender: sender,
	}
}

// Base route for currencies package
func (msg MsgDestroyCurrency) Route() string {
	return types.DefaultRoute
}

// Indeed type to destory currency
func (msg MsgDestroyCurrency) Type() string {
	return "destory_currency"
}

// Validate basic in case of destory message
func (msg MsgDestroyCurrency) ValidateBasic() sdk.Error {
	if msg.Sender.Empty() {
		return sdk.ErrInvalidAddress(msg.Sender.String())
	}

	if len(msg.Symbol) == 0 {
		return types.ErrWrongSymbol(msg.Symbol)
	}

	if msg.Amount.IsZero() {
		return types.ErrWrongAmount(msg.Amount.String())
	}

	// check denom, etc
	sdk.NewCoin(msg.Symbol, msg.Amount)

	return nil
}

// Get message bytes to sign
func (msg MsgDestroyCurrency) GetSignBytes() []byte {
	b, err := json.Marshal(msg)

	if err != nil {
		panic(err)
	}

	return sdk.MustSortJSON(b)
}

// Get signers for message
func (msg MsgDestroyCurrency) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}
