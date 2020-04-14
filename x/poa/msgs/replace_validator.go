// Message to replace validator described.
package msgs

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/dfinance/dnode/helpers"
	"github.com/dfinance/dnode/x/poa/types"
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
func (msg MsgReplaceValidator) ValidateBasic() error {
	if msg.OldValidator.Empty() {
		return sdkErrors.Wrap(sdkErrors.ErrInvalidAddress,"empty oldValidator")
	}

	if msg.NewValidator.Empty() {
		return sdkErrors.Wrap(sdkErrors.ErrInvalidAddress,"empty newValidator")
	}

	if len(msg.EthAddress) == 0 {
		return sdkErrors.Wrap(types.ErrWrongEthereumAddress,"empty")
	}

	if msg.Sender.Empty() {
		return sdkErrors.Wrap(sdkErrors.ErrInvalidAddress,"empty sender")
	}

	if !helpers.IsEthereumAddress(msg.EthAddress) {
		return sdkErrors.Wrapf(types.ErrWrongEthereumAddress, "%s for %s", msg.EthAddress, msg.NewValidator.String())
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
