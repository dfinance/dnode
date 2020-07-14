package types

import sdk "github.com/cosmos/cosmos-sdk/types"

const (
	EventTypeAdd    = ModuleName + ".add"
	EventTypeRemove = ModuleName + ".remove"
	//
	AttributeSdkAddress = "address"
	AttributeEthAddress = "eth_address"
)

// NewValidatorAddedEvent creates an Event on validator add (triggered on replace as well).
func NewValidatorAddedEvent(validator Validator) sdk.Event {
	return sdk.NewEvent(
		EventTypeAdd,
		sdk.Attribute{Key: AttributeSdkAddress, Value: validator.Address.String()},
		sdk.Attribute{Key: AttributeEthAddress, Value: validator.EthAddress},
	)
}

// NewValidatorRemovedEvent creates an Event on validator removal (triggered on replace as well).
func NewValidatorRemovedEvent(validator Validator) sdk.Event {
	return sdk.NewEvent(
		EventTypeRemove,
		sdk.Attribute{Key: AttributeSdkAddress, Value: validator.Address.String()},
		sdk.Attribute{Key: AttributeEthAddress, Value: validator.EthAddress},
	)
}
