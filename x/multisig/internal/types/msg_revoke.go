package types

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	dnTypes "github.com/dfinance/dnode/helpers/types"
)

// Client message to revoke an existing call confirm.
type MsgRevokeConfirm struct {
	// Call ID
	CallID dnTypes.ID `json:"call_id" yaml:"call_id" example:"0" format:"string representation for big.Uint" swaggertype:"string"`
	// Message sender address
	Sender sdk.AccAddress `json:"sender" yaml:"sender"`
}

// Implements sdk.Msg interface.
func (msg MsgRevokeConfirm) Route() string {
	return RouterKey
}

// Implements sdk.Msg interface.
func (msg MsgRevokeConfirm) Type() string {
	return "revoke_confirm"
}

// Implements sdk.Msg interface.
func (msg MsgRevokeConfirm) ValidateBasic() error {
	if err := msg.CallID.Valid(); err != nil {
		return sdkErrors.Wrap(ErrWrongCallId, err.Error())
	}

	if msg.Sender.Empty() {
		return sdkErrors.Wrap(sdkErrors.ErrInvalidAddress, "sender: empty")
	}

	return nil
}

// Implements sdk.Msg interface.
func (msg MsgRevokeConfirm) GetSignBytes() []byte {
	bc, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}

	return sdk.MustSortJSON(bc)
}

// Implements sdk.Msg interface.
func (msg MsgRevokeConfirm) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}

// NewMsgRevokeConfirm creates a new MsgRevokeConfirm message.
func NewMsgRevokeConfirm(callID dnTypes.ID, sender sdk.AccAddress) MsgRevokeConfirm {
	return MsgRevokeConfirm{
		CallID: callID,
		Sender: sender,
	}
}
