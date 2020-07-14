package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	EventTypeOrderPost            = ModuleName + ".post"
	EventTypeOrderCancel          = ModuleName + ".cancel"
	EventTypeFullyFilledOrder     = ModuleName + ".full_fill"
	EventTypePartiallyFilledOrder = ModuleName + ".partial_fill"
	//
	AttributeMarketId  = "market_id"
	AttributeOrderId   = "order_id"
	AttributeOwner     = "owner"
	AttributeDirection = "direction"
	AttributePrice     = "price"
	AttributeQuantity  = "quantity"
)

// NewOrderPostedEvent creates an Event on order post (creation).
func NewOrderPostedEvent(order Order) sdk.Event {
	return sdk.NewEvent(
		EventTypeOrderPost,
		sdk.NewAttribute(AttributeOwner, order.Owner.String()),
		sdk.NewAttribute(AttributeMarketId, order.Market.ID.String()),
		sdk.NewAttribute(AttributeOrderId, order.ID.String()),
		sdk.NewAttribute(AttributeDirection, order.Direction.String()),
		sdk.NewAttribute(AttributePrice, order.Price.String()),
		sdk.NewAttribute(AttributeQuantity, order.Quantity.String()),
	)
}

// NewOrderCanceledEvent creates an Event on order cancel (revoke / TTL).
func NewOrderCanceledEvent(order Order) sdk.Event {
	return sdk.NewEvent(
		EventTypeOrderCancel,
		sdk.NewAttribute(AttributeOwner, order.Owner.String()),
		sdk.NewAttribute(AttributeMarketId, order.Market.ID.String()),
		sdk.NewAttribute(AttributeOrderId, order.ID.String()),
		sdk.NewAttribute(AttributeDirection, order.Direction.String()),
		sdk.NewAttribute(AttributePrice, order.Price.String()),
		sdk.NewAttribute(AttributeQuantity, order.Quantity.String()),
	)
}

// NewFullyFilledOrderEvent creates an Event on order fully filled (triggered by Matcher).
func NewFullyFilledOrderEvent(order Order) sdk.Event {
	return sdk.NewEvent(
		EventTypeFullyFilledOrder,
		sdk.NewAttribute(AttributeOwner, order.Owner.String()),
		sdk.NewAttribute(AttributeMarketId, order.Market.ID.String()),
		sdk.NewAttribute(AttributeOrderId, order.ID.String()),
		sdk.NewAttribute(AttributeDirection, order.Direction.String()),
		sdk.NewAttribute(AttributePrice, order.Price.String()),
		sdk.NewAttribute(AttributeQuantity, order.Quantity.String()),
	)
}

// NewPartiallyFilledOrderEvent creates an Event on order partially filled (triggered by Matcher).
func NewPartiallyFilledOrderEvent(order Order) sdk.Event {
	return sdk.NewEvent(
		EventTypePartiallyFilledOrder,
		sdk.NewAttribute(AttributeOwner, order.Owner.String()),
		sdk.NewAttribute(AttributeMarketId, order.Market.ID.String()),
		sdk.NewAttribute(AttributeOrderId, order.ID.String()),
		sdk.NewAttribute(AttributeDirection, order.Direction.String()),
		sdk.NewAttribute(AttributePrice, order.Price.String()),
		sdk.NewAttribute(AttributeQuantity, order.Quantity.String()),
	)
}
