package msgs

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"encoding/json"
	types "wings-blockchain/x/currencies/types"
)

// Msg struct for issue new currencies
type MsgIssueCurrency struct {
	Symbol   string
	Amount   int64
	Decimals int8
	Creator  sdk.AccAddress
}

// Create new issue currency message
func NewMsgIssueCurrency(symbol string, supply int64, amount int8, creator sdk.AccAddress) MsgIssueCurrency {
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
func (msg MsgIssueCurrency) ValidateBasic() sdk.Error {
	if msg.Creator.Empty() {
		return sdk.ErrInvalidAddress(msg.Creator.String())
	}

	if len(msg.Symbol) == 0 {
		return types.ErrWrongSymbol(msg.Symbol)
	}

	if msg.Decimals < 0 || msg.Decimals > 8 || msg.Amount <= 0 {
		return types.ErrWrongDecimals(msg.Decimals)
	}

	return nil
}

// Getting bytes for signature
func (msg MsgIssueCurrency) GetSignBytes() []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}

	return sdk.MustSortJSON(b)
}

// Check who should sign message
func (msg MsgIssueCurrency) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Creator}
}
