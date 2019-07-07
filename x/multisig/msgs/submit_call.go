package msgs

import (
	types "wings-blockchain/x/multisig/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"encoding/json"
)

// Message for submit call
type MsgSubmitCall struct {
	Msg    types.MsMsg	  `json:"msg"`
	Sender sdk.AccAddress `json:"sender"`
}

func NewMsgSubmitCall(msg types.MsMsg, sender sdk.AccAddress) MsgSubmitCall {
	return MsgSubmitCall{
		Msg: msg,
		Sender: sender,
	}
}

func (msg MsgSubmitCall) Route() string {
	return types.DefaultRoute
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
