package types

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Client multisig message to replace an old PoA validator with a new one.
type MsgReplaceValidator struct {
	// Validator SDK address to remove
	OldValidator sdk.AccAddress `json:"old_validator" yaml:"old_validator"`
	// New validator SDK address
	NewValidator sdk.AccAddress `json:"new_validator" yaml:"new_validator"`
	// New validator Ethereum address
	EthAddress string `json:"eth_address" yaml:"eth_address"`
	// Message sender
	Sender sdk.AccAddress `json:"sender" yaml:"sender"`
}

// Implements sdk.Msg interface.
func (msg MsgReplaceValidator) Route() string {
	return RouterKey
}

// Implements sdk.Msg interface.
func (msg MsgReplaceValidator) Type() string {
	return "replace_validator"
}

// Implements sdk.Msg interface.
func (msg MsgReplaceValidator) ValidateBasic() error {
	if msg.OldValidator.Empty() {
		return sdkErrors.Wrap(sdkErrors.ErrInvalidAddress, "oldValidator: empty")
	}

	v := NewValidator(msg.NewValidator, msg.EthAddress)
	if err := v.Validate(); err != nil {
		return sdkErrors.Wrap(err, "newValidator")
	}

	if msg.OldValidator.Equals(msg.NewValidator) {
		return sdkErrors.Wrap(sdkErrors.ErrInvalidAddress, "oldValidator / newValidator: equal")
	}

	if msg.Sender.Empty() {
		return sdkErrors.Wrap(sdkErrors.ErrInvalidAddress, "sender: empty")
	}

	return nil
}

// Implements sdk.Msg interface.
func (msg MsgReplaceValidator) GetSignBytes() []byte {
	b, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}

	return sdk.MustSortJSON(b)
}

// Implements sdk.Msg interface.
func (msg MsgReplaceValidator) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}

// NewMsgReplaceValidator creates a new MsgReplaceValidator message.
func NewMsgReplaceValidator(oldValidator sdk.AccAddress, newValidator sdk.AccAddress, ethAddress string, sender sdk.AccAddress) MsgReplaceValidator {
	return MsgReplaceValidator{
		OldValidator: oldValidator,
		NewValidator: newValidator,
		EthAddress:   ethAddress,
		Sender:       sender,
	}
}
