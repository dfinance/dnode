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
	ModuleName        = types.ModuleName
	DefaultParamspace = types.DefaultParamspace
	// Event types, attribute types and values
	EventTypeCreate = types.EventTypeCreate
	//
	AttributeMarketId   = types.AttributeMarketId
	AttributeBaseDenom  = types.AttributeBaseDenom
	AttributeQuoteDenom = types.AttributeQuoteDenom
)

var (
	// variable aliases
	ModuleCdc = types.ModuleCdc
	// function aliases
	RegisterCodec       = types.RegisterCodec
	NewGenesisState     = types.NewGenesisState
	DefaultGenesisState = types.DefaultGenesisState
	ValidateGenesis     = types.ValidateGenesis
	DefaultParams       = types.DefaultParams
	NewMarket           = types.NewMarket
	NewMarketsFilter    = types.NewMarketsFilter
	NewMarketExtended   = types.NewMarketExtended
	NewKeeper           = keeper.NewKeeper
	NewQuerier          = keeper.NewQuerier
	// error aliases
	ErrWrongID         = types.ErrWrongID
	ErrWrongAssetDenom = types.ErrWrongAssetDenom
	ErrMarketExists    = types.ErrMarketExists
	ErrInvalidQuantity = types.ErrInvalidQuantity
	ErrWrongFrom       = types.ErrWrongFrom
)
