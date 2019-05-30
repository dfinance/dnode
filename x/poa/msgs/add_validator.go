package msgs

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"encoding/json"
	"wings-blockchain/x/poa/types"
)

const (
	MsgAddValidatorType = types.ModuleName + "/add-validator"
)

type MsgAddValidator struct {
	Address		sdk.AccAddress
	EthAddress 	string
	Sender		sdk.AccAddress
}

func NewMsgAddValidator(address sdk.AccAddress, ethAddress string) MsgAddValidator {
	return MsgAddValidator{
		Address: 	address,
		EthAddress: ethAddress,
	}
}

func (msg MsgAddValidator) Route() string {
	return types.DefaultRoute
}

func (msg MsgAddValidator) Type() string {
	return "add_validator"
}

func (msg MsgAddValidator) ValidateBasic() sdk.Error {
	if msg.Address.Empty() {
		return sdk.ErrInvalidAddress(msg.Address.String())
	}

	if len(msg.EthAddress) == 0 {
		return sdk.ErrUnknownRequest("Wrong Ethereum address for validator")
	}

	if msg.Sender.Empty() {
		return sdk.ErrInvalidAddress(msg.Sender.String())
	}

	return nil
}

func (msg MsgAddValidator) GetSignBytes() []byte {
	b, err := json.Marshal(msg)

	if err != nil {
		panic(err)
	}

	return sdk.MustSortJSON(b)
}

func (msg MsgAddValidator) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}
