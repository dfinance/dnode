// Implements message type to confirm call.
package msgs

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/dfinance/dnode/x/multisig/types"
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

func (msg MsgConfirmCall) ValidateBasic() error {
	if msg.Sender.Empty() {
		return sdkErrors.Wrap(sdkErrors.ErrInvalidAddress, "empty sender")
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
