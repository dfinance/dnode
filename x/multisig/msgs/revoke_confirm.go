// Implements new message type to revoke confirmation from call.
package msgs

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/WingsDao/wings-blockchain/x/multisig/types"
)

// Message to revoke confirmation from call.
type MsgRevokeConfirm struct {
	MsgId  uint64         `json:"msg_id"`
	Sender sdk.AccAddress `json:"sender"`
}

// Create new message instance to revoke confirmation.
func NewMsgRevokeConfirm(msgId uint64, sender sdk.AccAddress) MsgRevokeConfirm {
	return MsgRevokeConfirm{
		MsgId:  msgId,
		Sender: sender,
	}
}

func (msg MsgRevokeConfirm) Route() string {
	return types.RouterKey
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
