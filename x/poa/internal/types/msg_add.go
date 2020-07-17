package types

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Client multisig message to add a PoA validator.
type MsgAddValidator struct {
	// Validator SDK address
	Address sdk.AccAddress `json:"address" yaml:"address"`
	// Validator Ethereum address
	EthAddress string `json:"eth_address" yaml:"eth_address"`
	// Message sender
	Sender sdk.AccAddress `json:"sender" yaml:"sender"`
}

// Implements sdk.Msg interface.
func (msg MsgAddValidator) Route() string {
	return RouterKey
}

// Implements sdk.Msg interface.
func (msg MsgAddValidator) Type() string {
	return "add_validator"
}

// Implements sdk.Msg interface.
func (msg MsgAddValidator) ValidateBasic() error {
	v := NewValidator(msg.Address, msg.EthAddress)
	if err := v.Validate(); err != nil {
		return err
	}

	if msg.Sender.Empty() {
		return sdkErrors.Wrap(sdkErrors.ErrInvalidAddress, "sender: empty")
	}

	return nil
}

// Implements sdk.Msg interface.
func (msg MsgAddValidator) GetSignBytes() []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}

	return sdk.MustSortJSON(b)
}

// Implements sdk.Msg interface.
func (msg MsgAddValidator) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}

// NewMsgAddValidator creates a new MsgAddValidator message.
func NewMsgAddValidator(address sdk.AccAddress, ethAddress string, sender sdk.AccAddress) MsgAddValidator {
	return MsgAddValidator{
		Address:    address,
		EthAddress: ethAddress,
		Sender:     sender,
	}
}
