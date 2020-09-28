package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Client multisig message to unstake currency.
type MsgUnstakeCurrency struct {
	// Issue unique ID (could be txHash of transaction in another blockchain)
	ID string `json:"id" yaml:"id"`
	// Payee account (who unstakes)
	Staker sdk.AccAddress `json:"staker" yaml:"staker"`
}

// Implements sdk.Msg interface.
func (msg MsgUnstakeCurrency) Route() string {
	return RouterKey
}

// Implements sdk.Msg interface.
func (msg MsgUnstakeCurrency) Type() string {
	return "unstake_currency"
}

// Implements sdk.Msg interface.
func (msg MsgUnstakeCurrency) ValidateBasic() error {
	if len(msg.ID) == 0 {
		return sdkErrors.Wrap(ErrWrongIssueID, "empty")
	}

	if msg.Staker.Empty() {
		return sdkErrors.Wrap(sdkErrors.ErrInvalidAddress, "staker: empty")
	}

	return nil
}

// Implements sdk.Msg interface.
func (msg MsgUnstakeCurrency) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// Implements sdk.Msg interface.
// Msg is a multisig, so there are not signers.
func (msg MsgUnstakeCurrency) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{}
}

// NewMsgUnstakeCurrency creates a new MsgUnstakeCurrency message.
func NewMsgUnstakeCurrency(id string, staker sdk.AccAddress) MsgUnstakeCurrency {
	return MsgUnstakeCurrency{
		ID:     id,
		Staker: staker,
	}
}
