package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	dnTypes "github.com/dfinance/dnode/helpers/types"
)

// Client message to reduce currency balance.
type MsgWithdrawCurrency struct {
	// Target currency withdraw coin
	Coin sdk.Coin `json:"coin" yaml:"coin"`
	// Payer account (whose balance is decreased)
	Payer sdk.AccAddress `json:"payer" yaml:"payer"`
	// Second blockchain: payee account (whose balance is increased)
	PegZonePayee string `json:"pegzone_payee" yaml:"pegzone_payee"`
	// Second blockchain: ID
	PegZoneChainID string `json:"pegzone_chain_id" yaml:"pegzone_chain_id"`
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

	if msg.Payer.Empty() {
		return sdkErrors.Wrap(sdkErrors.ErrInvalidAddress, "payer: empty")
	}

	if msg.PegZonePayee == "" {
		return sdkErrors.Wrap(ErrWrongPegZonePayee, "empty")
	}

	return nil
}

// Implements sdk.Msg interface.
func (msg MsgWithdrawCurrency) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// Implements sdk.Msg interface.
func (msg MsgWithdrawCurrency) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Payer}
}

// NewMsgWithdrawCurrency creates a new MsgWithdrawCurrency message.
func NewMsgWithdrawCurrency(coin sdk.Coin, payer sdk.AccAddress, pzPayee, pzChainID string) MsgWithdrawCurrency {
	return MsgWithdrawCurrency{
		Coin:           coin,
		Payer:          payer,
		PegZonePayee:   pzPayee,
		PegZoneChainID: pzChainID,
	}
}
