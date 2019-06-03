package msgs

import (
	"wings-blockchain/x/multisig/types"
	"encoding/json"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type MsgRevokeConfirm struct {
	MsgId  uint64
	Sender sdk.AccAddress
}

func NewMsgRevokeConfirm(msgId uint64, sender sdk.AccAddress) MsgRevokeConfirm {
	return MsgRevokeConfirm{
		MsgId:  msgId,
		Sender: sender,
	}
}

func (msg MsgRevokeConfirm) Route() string {
	return types.DefaultRoute
}

func (msg MsgRevokeConfirm) Type() string {
	return "revoke_confirm"
}

func (msg MsgRevokeConfirm) ValidateBasic() sdk.Error {
	if msg.Sender.Empty() {
		return sdk.ErrInvalidAddress(msg.Sender.String())
	}

	return nil
}

func (msg MsgRevokeConfirm) GetSignBytes() []byte {
	bc, err := json.Marshal(msg)

	if err != nil {
		panic(err)
	}

	return sdk.MustSortJSON(bc)
}

func (msg MsgRevokeConfirm) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}
