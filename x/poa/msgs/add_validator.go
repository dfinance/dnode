// Message to add validator described.
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
	MsgAddValidatorType = types.ModuleName + "/add-validator"
)

// Message for adding validator
type MsgAddValidator struct {
	Address    sdk.AccAddress `json:"address"`
	EthAddress string         `json:"eth_address"`
	Sender     sdk.AccAddress `json:"sender"`
}

// Create new 'add validator' message
func NewMsgAddValidator(address sdk.AccAddress, ethAddress string, sender sdk.AccAddress) MsgAddValidator {
	return MsgAddValidator{
		Address:    address,
		EthAddress: ethAddress,
		Sender:     sender,
	}
}

// Message route
func (msg MsgAddValidator) Route() string {
	return types.RouterKey
}

// Message type
func (msg MsgAddValidator) Type() string {
	return "add_validator"
}

// Validate basic for add validator msg
func (msg MsgAddValidator) ValidateBasic() error {
	if msg.Address.Empty() {
		return sdkErrors.Wrap(sdkErrors.ErrInvalidAddress,"empty address")
	}

	if len(msg.EthAddress) == 0 {
		return sdkErrors.Wrap(types.ErrWrongEthereumAddress, "empty")
	}

	if msg.Sender.Empty() {
		return sdkErrors.Wrap(sdkErrors.ErrInvalidAddress, "empty sender")
	}

	if !helpers.IsEthereumAddress(msg.EthAddress) {
		return sdkErrors.Wrapf(types.ErrWrongEthereumAddress, "%s for %s", msg.EthAddress, msg.Address.String())
	}

	return nil
}

// Get signature bytes
func (msg MsgAddValidator) GetSignBytes() []byte {
	b, err := json.Marshal(msg)

	if err != nil {
		panic(err)
	}

	return sdk.MustSortJSON(b)
}

// Get signers addresses
func (msg MsgAddValidator) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Sender}
}
