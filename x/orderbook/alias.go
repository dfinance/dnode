package orderbook

import (
	"github.com/dfinance/dnode/x/orderbook/internal/keeper"
	"github.com/dfinance/dnode/x/orderbook/internal/types"
)

type (
	Keeper       = keeper.Keeper
	HistoryItem  = types.HistoryItem
	HistoryItems = types.HistoryItems
)

const (
	ModuleName = types.ModuleName
	StoreKey   = types.StoreKey
)

var (
	// variable aliases
	ModuleCdc = types.ModuleCdc
	// function aliases
	RegisterCodec     = types.RegisterCodec
	NewHistoryItem    = types.NewHistoryItem
	NewClearanceEvent = types.NewClearanceEvent
	NewKeeper         = keeper.NewKeeper
	NewMatcherPool    = keeper.NewMatcherPool
)
