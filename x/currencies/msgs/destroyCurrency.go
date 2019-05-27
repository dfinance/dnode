package msgs

import (
	"github.com/cosmos/cosmos-sdk/types"
	"encoding/json"
)

// Message for destory currency
type MsgDestroyCurrency struct {
	Symbol 	 string
	Amount   int64
	Sender   types.AccAddress
}

// Create new message to destory currency
func NewMsgDestroyCurrency(symbol string, amount int64, sender types.AccAddress) MsgDestroyCurrency {
	return MsgDestroyCurrency{
		Symbol: symbol,
		Amount: amount,
		Sender: sender,
	}
}

// Base route for currencies package
func (msg MsgDestroyCurrency) Route() string {
	return "currencies"
}

// Indeed type to destory currency
func (msg MsgDestroyCurrency) Type() string {
	return "destory_currency"
}

// Validate basic in case of destory message
func (msg MsgDestroyCurrency) ValidateBasic() types.Error {
	if msg.Sender.Empty() {
		return types.ErrInvalidAddress(msg.Sender.String())
	}

	if len(msg.Symbol) == 0 {
		return types.ErrUnknownRequest("Symbol should be not empty")
	}

	if msg.Amount == 0 {
		return types.ErrUnknownRequest("amount can't be less/equal 0")
	}

	return nil
}

// Get message bytes to sign
func (msg MsgDestroyCurrency) GetSignBytes() []byte {
	b, err := json.Marshal(msg)

	if err != nil {
		panic(err)
	}

	return types.MustSortJSON(b)
}

// Get signers for message
func (msg MsgDestroyCurrency) GetSigners() []types.AccAddress {
	return []types.AccAddress{msg.Sender}
}
