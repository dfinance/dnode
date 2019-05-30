package msgs

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"encoding/json"
	types "wings-blockchain/x/poa/types"
)

const (
	MsgRemoveValidatorType = types.ModuleName + "/remove-validator"
)

type MsgRemoveValidator struct {
	Address sdk.AccAddress
	Sender  sdk.AccAddress
}

func NewMsgRemoveValidator(address sdk.AccAddress, sender sdk.AccAddress) MsgRemoveValidator {
	return MsgRemoveValidator{
		Address: address,
		Sender:  sender,
	}
}

func (msg MsgRemoveValidator) Route() string {
	return types.DefaultRoute
}

func (msg MsgRemoveValidator) Type() string {
	return "remove_validator"
}

func (msg MsgRemoveValidator) ValidateBasic() sdk.Error {
	if msg.Address.Empty() {
		return sdk.ErrInvalidAddress(msg.Address.String())
	}

	if msg.Sender.Empty() {
		return sdk.ErrInvalidAddress(msg.Sender.String())
	}

	return nil
}

func (msg MsgRemoveValidator) GetSignBytes() []byte {
	b, err := json.Marshal(msg)

	if err != nil {
		panic(err)
	}

	return sdk.MustSortJSON(b)
}

func (msg MsgRemoveValidator) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}