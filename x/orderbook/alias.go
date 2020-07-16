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
	// Event types, attribute types and values
	EventTypeClearance = types.EventTypeClearance
	//
	AttributeMarketId = types.AttributeMarketId
	AttributePrice    = types.AttributePrice
	// Permissions
	PermHistoryReader = types.PermHistoryReader
	PermHistoryWriter = types.PermHistoryWriter
	PermOrdersRead    = types.PermOrdersRead
	PermExecFills     = types.PermExecFills
)

var (
	// variable aliases
	ModuleCdc            = types.ModuleCdc
	AvailablePermissions = types.AvailablePermissions
	// function aliases
	RegisterCodec     = types.RegisterCodec
	NewHistoryItem    = types.NewHistoryItem
	NewClearanceEvent = types.NewClearanceEvent
	NewKeeper         = keeper.NewKeeper
	NewMatcherPool    = keeper.NewMatcherPool
	// perms requests
	RequestOrdersPerms = types.RequestOrdersPerms
)
