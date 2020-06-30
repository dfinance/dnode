package types

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	dnTypes "github.com/dfinance/dnode/helpers/types"
)

// Client message to reduce currency balance.
type MsgWithdrawCurrency struct {
	// Target currency denom
	Denom string `json:"denom"`
	// Withdraw amount
	Amount sdk.Int `json:"amount"`
	// Target account
	Spender sdk.AccAddress `json:"spender"`
	// Second blockchain: spender account
	PegZoneRecipient string `json:"pregzone_spender"`
	// Second blockchain: ID
	PegZoneChainID string `json:"pegzone_chainID"`
}

// Implements sdk.Msg interface.
func (msg MsgWithdrawCurrency) Route() string {
	return RouterKey
}

// Implements sdk.Msg interface.
func (msg MsgWithdrawCurrency) Type() string {
	return "withdraw_currency"
}

// Implements sdk.Msg interface.
func (msg MsgWithdrawCurrency) ValidateBasic() error {
	if err := dnTypes.DenomFilter(msg.Denom); err != nil {
		return sdkErrors.Wrap(ErrWrongDenom, err.Error())
	}

	if msg.Amount.IsZero() {
		return sdkErrors.Wrap(ErrWrongAmount, "zero")
	}

	if msg.Spender.Empty() {
		return sdkErrors.Wrap(sdkErrors.ErrInvalidAddress, "spender: empty")
	}

	if len(msg.PegZoneRecipient) == 0 {
		return sdkErrors.Wrap(ErrWrongPegZoneSpender, "empty")
	}

	// check sdk.Coin is creatable
	sdk.NewCoin(msg.Denom, msg.Amount)

	return nil
}

// Implements sdk.Msg interface.
func (msg MsgWithdrawCurrency) GetSignBytes() []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}

	return sdk.MustSortJSON(b)
}

// Implements sdk.Msg interface.
func (msg MsgWithdrawCurrency) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Spender}
}

// NewMsgWithdrawCurrency creates a new MsgWithdrawCurrency message.
func NewMsgWithdrawCurrency(denom string, amount sdk.Int, spender sdk.AccAddress, pzSpender, pzChainID string) MsgWithdrawCurrency {
	return MsgWithdrawCurrency{
		Denom:            denom,
		Amount:           amount,
		Spender:          spender,
		PegZoneRecipient: pzSpender,
		PegZoneChainID:   pzChainID,
	}
}
