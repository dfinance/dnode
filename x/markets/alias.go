package markets

import (
	"github.com/dfinance/dnode/x/markets/internal/keeper"
	"github.com/dfinance/dnode/x/markets/internal/types"
)

type (
	Keeper          = keeper.Keeper
	Market          = types.Market
	Markets         = types.Markets
	MarketExtended  = types.MarketExtended
	MsgCreateMarket = types.MsgCreateMarket
	GenesisState    = types.GenesisState
)

const (
	ModuleName = types.ModuleName
	StoreKey   = types.StoreKey
	// Event types, attribute types and values
	EventTypeCreate = types.EventTypeCreate
	//
	AttributeMarketId   = types.AttributeMarketId
	AttributeBaseDenom  = types.AttributeBaseDenom
	AttributeQuoteDenom = types.AttributeQuoteDenom
)

var (
	// variable aliases
	ModuleCdc            = types.ModuleCdc
	AvailablePermissions = types.AvailablePermissions
	// function aliases
	RegisterCodec       = types.RegisterCodec
	NewKeeper           = keeper.NewKeeper
	NewQuerier          = keeper.NewQuerier
	DefaultGenesisState = types.DefaultGenesisState
	NewMarket           = types.NewMarket
	NewMarketsFilter    = types.NewMarketsFilter
	NewMarketExtended   = types.NewMarketExtended
	// perms requests
	RequestCCStoragePerms = types.RequestCCStoragePerms
	// error aliases
	ErrWrongID         = types.ErrWrongID
	ErrWrongAssetDenom = types.ErrWrongAssetDenom
	ErrMarketExists    = types.ErrMarketExists
	ErrInvalidQuantity = types.ErrInvalidQuantity
	ErrWrongFrom       = types.ErrWrongFrom
)
