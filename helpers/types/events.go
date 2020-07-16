package types

import sdk "github.com/cosmos/cosmos-sdk/types"

func NewModuleNameEvent(name string) sdk.Event {
	return sdk.NewEvent(
		sdk.EventTypeMessage,
		sdk.NewAttribute(sdk.AttributeKeyModule, name),
	)
}
