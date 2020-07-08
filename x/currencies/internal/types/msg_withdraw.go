package types

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	dnTypes "github.com/dfinance/dnode/helpers/types"
)

// Client message to reduce currency balance.
type MsgWithdrawCurrency struct {
	// Target currency withdraw coin
	Coin sdk.Coin `json:"coin"`
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
	if err := dnTypes.DenomFilter(msg.Coin.Denom); err != nil {
		return sdkErrors.Wrap(ErrWrongDenom, err.Error())
	}

	if msg.Coin.Amount.LTE(sdk.ZeroInt()) {
		return sdkErrors.Wrap(ErrWrongAmount, "LTE to zero")
	}

	if msg.Spender.Empty() {
		return sdkErrors.Wrap(sdkErrors.ErrInvalidAddress, "spender: empty")
	}

	if len(msg.PegZoneRecipient) == 0 {
		return sdkErrors.Wrap(ErrWrongPegZoneSpender, "empty")
	}

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
func NewMsgWithdrawCurrency(coin sdk.Coin, spender sdk.AccAddress, pzSpender, pzChainID string) MsgWithdrawCurrency {
	return MsgWithdrawCurrency{
		Coin:             coin,
		Spender:          spender,
		PegZoneRecipient: pzSpender,
		PegZoneChainID:   pzChainID,
	}
}
