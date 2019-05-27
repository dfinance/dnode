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
func NewMessageDestoryCurrency(symbol string, amount int64, sender types.AccAddress) MsgDestroyCurrency {
	return MsgDestroyCurrency{
		Symbol: symbol,
		Amount: amount,
		Sender: sender,
	}
}

// Base route for currencies
func (msg MsgDestroyCurrency) Route() string {
	return "currencies"
}

func (msg MsgDestroyCurrency) Type() string {
	return "destory-currency"
}

func (msg MsgDestroyCurrency) GetSignBytes() []byte {
	b, err := json.Marshal(msg)

	if err != nil {
		panic(err)
	}

	return types.MustSortJSON(b)
}

func (msg MsgDestroyCurrency) GetSigners() []types.AccAddress {
	return []types.AccAddress{msg.Sender}
}

