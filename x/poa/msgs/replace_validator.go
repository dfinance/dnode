// Message to replace validator described.
package msgs

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/WingsDao/wings-blockchain/helpers"
	"github.com/WingsDao/wings-blockchain/x/poa/types"
)

// Type for codec
const (
	MsgReplaceValidatorType = types.ModuleName + "/replace-validator"
)

// Message for replace validator
type MsgReplaceValidator struct {
	OldValidator sdk.AccAddress `json:"old_address"`
	NewValidator sdk.AccAddress `json:"new_validator"`
	EthAddress   string         `json:"eth_address"`
	Sender       sdk.AccAddress `json:"sender"`
}

// Create new 'replace validator' message
func NewMsgReplaceValidator(oldValidator sdk.AccAddress, newValidator sdk.AccAddress, ethAddress string, sender sdk.AccAddress) MsgReplaceValidator {
	return MsgReplaceValidator{
		OldValidator: oldValidator,
		NewValidator: newValidator,
		EthAddress:   ethAddress,
		Sender:       sender,
	}
}

// Message route
func (msg MsgReplaceValidator) Route() string {
	return types.RouterKey
}

// Message type
func (msg MsgReplaceValidator) Type() string {
	return "replace_validator"
}

// Validate basic 'replace validator' message
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

	if !helpers.IsEthereumAddress(msg.EthAddress) {
		return types.ErrWrongEthereumAddress(msg.EthAddress)
	}

	return nil
}

// Get bytes to sign from message
func (msg MsgReplaceValidator) GetSignBytes() []byte {
	b, err := json.Marshal(msg)

	if err != nil {
		panic(err)
	}

	return sdk.MustSortJSON(b)
}

// Get signers addresses
func (msg MsgReplaceValidator) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}
