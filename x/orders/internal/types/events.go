package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	dnTypes "github.com/dfinance/dnode/helpers/types"
)

const (
	EventTypeOrderPost            = ModuleName + ".post"
	EventTypeOrderCancel          = ModuleName + ".cancel"
	EventTypeFullyFilledOrder     = ModuleName + ".full_fill"
	EventTypePartiallyFilledOrder = ModuleName + ".partial_fill"
	EventAttributeKeyOwner        = "owner"
	EventAttributeKeyOrderID      = "order_id"
	EventAttributeKeyMarketID     = "market_id"
)

func NewOrderPostedEvent(owner sdk.AccAddress, marketID, orderID dnTypes.ID) sdk.Event {
	return sdk.NewEvent(
		EventTypeOrderPost,
		sdk.NewAttribute(dnTypes.DnEventAttrKey, dnTypes.DnEventAttrValue),
		sdk.NewAttribute(EventAttributeKeyOwner, owner.String()),
		sdk.NewAttribute(EventAttributeKeyMarketID, marketID.String()),
		sdk.NewAttribute(EventAttributeKeyOrderID, orderID.String()),
	)
}

func NewOrderCanceledEvent(owner sdk.AccAddress, marketID, orderID dnTypes.ID) sdk.Event {
	return sdk.NewEvent(
		EventTypeOrderCancel,
		sdk.NewAttribute(dnTypes.DnEventAttrKey, dnTypes.DnEventAttrValue),
		sdk.NewAttribute(EventAttributeKeyOwner, owner.String()),
		sdk.NewAttribute(EventAttributeKeyMarketID, marketID.String()),
		sdk.NewAttribute(EventAttributeKeyOrderID, orderID.String()),
	)
}

func NewFullyFilledOrderEvent(owner sdk.AccAddress, marketID, orderID dnTypes.ID) sdk.Event {
	return sdk.NewEvent(
		EventTypeFullyFilledOrder,
		sdk.NewAttribute(dnTypes.DnEventAttrKey, dnTypes.DnEventAttrValue),
		sdk.NewAttribute(EventAttributeKeyOwner, owner.String()),
		sdk.NewAttribute(EventAttributeKeyMarketID, marketID.String()),
		sdk.NewAttribute(EventAttributeKeyOrderID, orderID.String()),
	)
}

func NewPartiallyFilledOrderEvent(owner sdk.AccAddress, marketID, orderID dnTypes.ID) sdk.Event {
	return sdk.NewEvent(
		EventTypePartiallyFilledOrder,
		sdk.NewAttribute(dnTypes.DnEventAttrKey, dnTypes.DnEventAttrValue),
		sdk.NewAttribute(EventAttributeKeyOwner, owner.String()),
		sdk.NewAttribute(EventAttributeKeyMarketID, marketID.String()),
		sdk.NewAttribute(EventAttributeKeyOrderID, orderID.String()),
	)
}
