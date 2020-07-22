package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	dnTypes "github.com/dfinance/dnode/helpers/types"
)

// Client message to confirm an existing call.
type MsgConfirmCall struct {
	// Confirming CallID
	CallID dnTypes.ID `json:"call_id" yaml:"call_id" example:"0" format:"string representation for big.Uint" swaggertype:"string"`
	// PoA validator address
	Sender sdk.AccAddress `json:"sender" yaml:"sender"`
}

// Implements sdk.Msg interface.
func (msg MsgConfirmCall) Route() string {
	return RouterKey
}

// Implements sdk.Msg interface.
func (msg MsgConfirmCall) Type() string {
	return "confirm_call"
}

// Implements sdk.Msg interface.
func (msg MsgConfirmCall) ValidateBasic() error {
	if err := msg.CallID.Valid(); err != nil {
		return sdkErrors.Wrap(ErrWrongCallId, err.Error())
	}

	if msg.Sender.Empty() {
		return sdkErrors.Wrap(sdkErrors.ErrInvalidAddress, "sender: empty")
	}

	return nil
}

// Implements sdk.Msg interface.
func (msg MsgConfirmCall) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// Implements sdk.Msg interface.
func (msg MsgConfirmCall) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}

// NewMsgConfirmCall creates a new MsgConfirmCall message.
func NewMsgConfirmCall(callID dnTypes.ID, sender sdk.AccAddress) MsgConfirmCall {
	return MsgConfirmCall{
		CallID: callID,
		Sender: sender,
	}
}
