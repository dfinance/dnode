package orders

import (
	"github.com/dfinance/dnode/x/orders/internal/keeper"
	"github.com/dfinance/dnode/x/orders/internal/types"
)

type (
	Keeper         = keeper.Keeper
	Order          = types.Order
	Orders         = types.Orders
	OrderFill      = types.OrderFill
	OrderFills     = types.OrderFills
	Direction      = types.Direction
	MsgPostOrder   = types.MsgPostOrder
	MsgRevokeOrder = types.MsgRevokeOrder
	OrdersReq      = types.OrdersReq
)

const (
	ModuleName   = types.ModuleName
	StoreKey     = types.StoreKey
	BidDirection = types.Bid
	AskDirection = types.Ask
	// Event types, attribute types
	EventTypeOrderPost            = types.EventTypeOrderPost
	EventTypeOrderCancel          = types.EventTypeOrderCancel
	EventTypeFullyFilledOrder     = types.EventTypeFullyFilledOrder
	EventTypePartiallyFilledOrder = types.EventTypePartiallyFilledOrder
	//
	AttributeKeyMarketID = types.AttributeMarketId
	AttributeKeyOrderID  = types.AttributeOrderId
	AttributeKeyOwner    = types.AttributeOwner
	AttributeKeyQuantity = types.AttributeQuantity
)

var (
	// variable aliases
	ModuleCdc = types.ModuleCdc
	// function aliases
	RegisterCodec = types.RegisterCodec
	NewKeeper     = keeper.NewKeeper
	NewQuerier    = keeper.NewQuerier
	// error aliases
	ErrWrongMarketID  = types.ErrWrongMarketID
	ErrWrongOwner     = types.ErrWrongOwner
	ErrWrongPrice     = types.ErrWrongPrice
	ErrWrongQuantity  = types.ErrWrongQuantity
	ErrWrongTtl       = types.ErrWrongTtl
	ErrWrongDirection = types.ErrWrongDirection
	ErrWrongOrderID   = types.ErrWrongOrderID
	ErrWrongAssetCode = types.ErrWrongAssetCode
)
