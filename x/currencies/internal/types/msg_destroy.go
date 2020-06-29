package types

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	dnTypes "github.com/dfinance/dnode/helpers/types"
)

// Client message to destroy currency.
type MsgDestroyCurrency struct {
	// Target currency denom
	Denom string `json:"denom"`
	// Destroy amount
	Amount sdk.Int `json:"amount"`
	// Target account
	Spender sdk.AccAddress `json:"spender"`
	// Second blockchain: spender account
	Recipient string `json:"recipient"`
	// Second blockchain: ID
	ChainID string `json:"chainID"`
}

// Implements sdk.Msg interface.
func (msg MsgDestroyCurrency) Route() string {
	return RouterKey
}

// Implements sdk.Msg interface.
func (msg MsgDestroyCurrency) Type() string {
	return "destroy_currency"
}

// Implements sdk.Msg interface.
func (msg MsgDestroyCurrency) ValidateBasic() error {
	if err := dnTypes.DenomFilter(msg.Denom); err != nil {
		return sdkErrors.Wrap(ErrWrongDenom, err.Error())
	}

	if msg.Amount.IsZero() {
		return sdkErrors.Wrap(ErrWrongAmount, "zero")
	}

	if msg.Spender.Empty() {
		return sdkErrors.Wrap(sdkErrors.ErrInvalidAddress, "spender: empty")
	}

	if len(msg.Recipient) == 0 {
		return sdkErrors.Wrap(ErrWrongRecipient, "empty")
	}

	// check sdk.Coin is creatable
	sdk.NewCoin(msg.Denom, msg.Amount)

	return nil
}

// Implements sdk.Msg interface.
func (msg MsgDestroyCurrency) GetSignBytes() []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}

	return sdk.MustSortJSON(b)
}

// Implements sdk.Msg interface.
func (msg MsgDestroyCurrency) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Spender}
}

// NewMsgDestroyCurrency creates a new MsgDestroyCurrency message.
func NewMsgDestroyCurrency(denom string, amount sdk.Int, spender sdk.AccAddress, recipient, chainID string) MsgDestroyCurrency {
	return MsgDestroyCurrency{
		Denom:     denom,
		Amount:    amount,
		Spender:   spender,
		Recipient: recipient,
		ChainID:   chainID,
	}
}
