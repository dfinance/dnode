package msgs

import (
	"github.com/cosmos/cosmos-sdk/types"
	"encoding/json"
)

// Msg struct for issue new currencies
type MsgIssueCurrency struct {
	Symbol   string
	Amount   int64
	Decimals int8
	Creator  types.AccAddress
}

// Create new issue currency message
func NewMsgIssueCurrency(symbol string, supply int64, amount int8, creator types.AccAddress) MsgIssueCurrency {
	return MsgIssueCurrency{
		Symbol:   symbol,
		Amount:   supply,
		Decimals: amount,
		Creator:  creator,
	}
}

// Common router for currencies package
func (msg MsgIssueCurrency) Route() string {
	return "currencies"
}

// Command for issue new currencies
func (msg MsgIssueCurrency) Type() string {
	return "issue_currency"
}

// Basic validation, without state
func (msg MsgIssueCurrency) ValidateBasic() types.Error {
	if msg.Creator.Empty() {
		return types.ErrInvalidAddress(msg.Creator.String())
	}

	if len(msg.Symbol) == 0 {
		return types.ErrUnknownRequest("Symbol should be not empty")
	}

	if msg.Decimals < 0 || msg.Decimals > 8 || msg.Amount <= 0 {
		return types.ErrUnknownRequest("Decimals or amount can't be less/equal 0, " +
			"and decimals should be less then 8")
	}

	return nil
}

// Getting bytes for signature
func (msg MsgIssueCurrency) GetSignBytes() []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}

	return types.MustSortJSON(b)
}

// Check who should sign message
func (msg MsgIssueCurrency) GetSigners() []types.AccAddress {
	return []types.AccAddress{msg.Creator}
}
