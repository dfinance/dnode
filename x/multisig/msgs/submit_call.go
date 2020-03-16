// Create new message type.
package msgs

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dfinance/dnode/x/core"
	types "github.com/dfinance/dnode/x/multisig/types"
)

// Message for submit call
type MsgSubmitCall struct {
	Msg      core.MsMsg     `json:"msg"`
	UniqueID string         `json:"uniqueID"`
	Sender   sdk.AccAddress `json:"sender"`
}

// Create new instance of message to submit call.
func NewMsgSubmitCall(msg core.MsMsg, uniqueID string, sender sdk.AccAddress) MsgSubmitCall {
	return MsgSubmitCall{
		Msg:      msg,
		UniqueID: uniqueID,
		Sender:   sender,
	}
}

func (msg MsgSubmitCall) Route() string {
	return types.RouterKey
}

func (msg MsgSubmitCall) Type() string {
	return "submit_call"
}

func (msg MsgSubmitCall) ValidateBasic() sdk.Error {
	err := msg.Msg.ValidateBasic()
	if err != nil {
		return err
	}

	if msg.Sender.Empty() {
		return sdk.ErrInvalidAddress(msg.Sender.String())
	}

	return nil
}

func (msg MsgSubmitCall) GetSignBytes() []byte {
	bc, err := json.Marshal(msg)

	if err != nil {
		panic(err)
	}

	return sdk.MustSortJSON(bc)
}

func (msg MsgSubmitCall) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}
