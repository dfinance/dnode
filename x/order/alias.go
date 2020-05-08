package order

import (
	"github.com/dfinance/dnode/x/order/internal/keeper"
	"github.com/dfinance/dnode/x/order/internal/types"
)

type (
	Keeper     = keeper.Keeper
	Order      = types.Order
	Orders     = types.Orders
	OrderFill  = types.OrderFill
	OrderFills = types.OrderFills
	Direction  = types.Direction
)

const (
	ModuleName   = types.ModuleName
	StoreKey     = types.StoreKey
	BidDirection = types.Bid
	AskDirection = types.Ask
)

var (
	// variable aliases
	ModuleCdc = types.ModuleCdc
	// function aliases
	NewKeeper = keeper.NewKeeper
)
