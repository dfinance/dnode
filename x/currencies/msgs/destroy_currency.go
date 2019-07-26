package msgs

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"encoding/json"
	"wings-blockchain/x/currencies/types"
)

// Message for destory currency
type MsgDestroyCurrency struct {
    ChainID  string         `json:"chainID"`
	Symbol 	 string			`json:"symbol"`
	Amount   sdk.Int	 	`json:"amount"`
	Spender  sdk.AccAddress `json:"spender"`
}

// Create new message to destory currency
func NewMsgDestroyCurrency(chainID string, symbol string, amount sdk.Int, spender sdk.AccAddress) MsgDestroyCurrency {
	return MsgDestroyCurrency{
	    ChainID: chainID,
		Symbol: symbol,
		Amount: amount,
		Spender: spender,
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
	if msg.Spender.Empty() {
		return sdk.ErrInvalidAddress(msg.Spender.String())
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
	return []sdk.AccAddress{msg.Spender}
}
