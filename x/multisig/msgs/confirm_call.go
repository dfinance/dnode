// Implements message type to confirm call.
package msgs

import (
	"encoding/json"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"wings-blockchain/x/multisig/types"
)

// Message type.
type MsgConfirmCall struct {
	MsgId  uint64         `json:"msg_id"`
	Sender sdk.AccAddress `json:"sender"`
}

// New instance of message.
func NewMsgConfirmCall(msgId uint64, sender sdk.AccAddress) MsgConfirmCall {
	return MsgConfirmCall{
		MsgId:  msgId,
		Sender: sender,
	}
}

func (msg MsgConfirmCall) Route() string {
	return types.RouterKey
}

func (msg MsgConfirmCall) Type() string {
	return "confirm_call"
}

func (msg MsgConfirmCall) ValidateBasic() sdk.Error {
	if msg.Sender.Empty() {
		return sdk.ErrInvalidAddress(msg.Sender.String())
	}

	return nil
}

func (msg MsgConfirmCall) GetSignBytes() []byte {
	bc, err := json.Marshal(msg)

	if err != nil {
		panic(err)
	}

	return sdk.MustSortJSON(bc)
}

func (msg MsgConfirmCall) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}
