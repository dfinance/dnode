package types

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	dnTypes "github.com/dfinance/dnode/helpers/types"
)

// Client message to confirm an existing call.
type MsgConfirmCall struct {
	// Confirming CallID
	CallID dnTypes.ID `json:"call_id"`
	// Message sender address
	Sender sdk.AccAddress `json:"sender"`
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
	bc, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}

	return sdk.MustSortJSON(bc)
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
