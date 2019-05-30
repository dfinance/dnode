package msgs

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"encoding/json"
	"wings-blockchain/x/poa/types"
)

const (
	MsgReplaceValidatorType = types.ModuleName + "/replace-validator"
)

type MsgReplaceValidator struct {
	OldValidator sdk.AccAddress
	NewValidator sdk.AccAddress
	EthAddress	 string
	Sender 		 sdk.AccAddress
}

func NewMsgReplaceValidator(oldValidator sdk.AccAddress, newValidator sdk.AccAddress, ethAddress string, sender sdk.AccAddress)  MsgReplaceValidator {
	return MsgReplaceValidator{
		OldValidator: oldValidator,
		NewValidator: newValidator,
		EthAddress:   ethAddress,
		Sender: 	  sender,
	}
}

func (msg MsgReplaceValidator) Route() string {
	return types.DefaultRoute
}

func (msg MsgReplaceValidator) Type() string {
	return "replace_validator"
}

func (msg MsgReplaceValidator) ValidateBasic() sdk.Error {
	if msg.OldValidator.Empty() {
		return sdk.ErrInvalidAddress(msg.OldValidator.String())
	}

	if msg.NewValidator.Empty() {
		return sdk.ErrInvalidAddress(msg.NewValidator.String())
	}

	if len(msg.EthAddress) == 0 {
		return sdk.ErrUnknownRequest("Wrong Ethereum address for validator")
	}

	if msg.Sender.Empty() {
		return sdk.ErrInvalidAddress(msg.Sender.String())
	}

	return nil
}

func (msg MsgReplaceValidator) GetSignBytes() []byte {
	b, err := json.Marshal(msg)

	if err != nil {
		panic(err)
	}

	return sdk.MustSortJSON(b)
}

func (msg MsgReplaceValidator) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}