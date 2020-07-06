package types

import (
	"encoding/json"

	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Client multisig message to remove a PoA validator.
type MsgRemoveValidator struct {
	// Validator SDK address
	Address sdk.AccAddress `json:"address"`
	// Message sender
	Sender sdk.AccAddress `json:"sender"`
}

// Implements sdk.Msg interface.
func (msg MsgRemoveValidator) Route() string {
	return RouterKey
}

// Implements sdk.Msg interface.
func (msg MsgRemoveValidator) Type() string {
	return "remove_validator"
}

// Implements sdk.Msg interface.
func (msg MsgRemoveValidator) ValidateBasic() error {
	if msg.Address.Empty() {
		return sdkErrors.Wrap(sdkErrors.ErrInvalidAddress, "address: empty")
	}

	if msg.Sender.Empty() {
		return sdkErrors.Wrap(sdkErrors.ErrInvalidAddress, "sender: empty")
	}

	return nil
}

// Implements sdk.Msg interface.
func (msg MsgRemoveValidator) GetSignBytes() []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}

	return sdk.MustSortJSON(b)
}

// Implements sdk.Msg interface.
func (msg MsgRemoveValidator) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}

// NewMsgRemoveValidator creates a new MsgRemoveValidator message.
func NewMsgRemoveValidator(address sdk.AccAddress, sender sdk.AccAddress) MsgRemoveValidator {
	return MsgRemoveValidator{
		Address: address,
		Sender:  sender,
	}
}
