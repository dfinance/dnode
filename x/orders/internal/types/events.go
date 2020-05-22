package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	dnTypes "github.com/dfinance/dnode/helpers/types"
)

const (
	EventTypeFullyFilledOrder = ModuleName + "/fully_filled"
	EventAttributeKeyOwner    = "owner"
	EventAttributeKeyOrderID  = "order_id"
	EventAttributeKeyQuantity = "quantity"
)

func NewFullyFilledOrderEvent(owner sdk.AccAddress, orderID dnTypes.ID) sdk.Event {
	return sdk.NewEvent(
		EventTypeFullyFilledOrder,
		sdk.NewAttribute(EventAttributeKeyOwner, owner.String()),
		sdk.NewAttribute(EventAttributeKeyOrderID, orderID.String()),
	)
}

func NewPartiallyFilledOrderEvent(owner sdk.AccAddress, orderID dnTypes.ID, quantity sdk.Uint) sdk.Event {
	return sdk.NewEvent(
		EventTypeFullyFilledOrder,
		sdk.NewAttribute(EventAttributeKeyOwner, owner.String()),
		sdk.NewAttribute(EventAttributeKeyOrderID, orderID.String()),
		sdk.NewAttribute(EventAttributeKeyQuantity, quantity.String()),
	)
}
