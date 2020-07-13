package types

import sdk "github.com/cosmos/cosmos-sdk/types"

const (
	DnEventAttrKey   = "dn_type"
	DnEventAttrValue = "yes"
)

func NewDnEventAttribute() sdk.Attribute {
	return sdk.Attribute{Key: DnEventAttrKey, Value: DnEventAttrValue}
}

func NewModuleNameEvent(name string) sdk.Event {
	return sdk.NewEvent(
		sdk.EventTypeMessage,
		sdk.NewAttribute(sdk.AttributeKeyModule, name),
	)
}
