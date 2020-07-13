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
