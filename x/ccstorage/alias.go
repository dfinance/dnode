package ccstorage

import (
	"github.com/dfinance/dnode/x/ccstorage/internal/keeper"
	"github.com/dfinance/dnode/x/ccstorage/internal/types"
)

type (
	Keeper          = keeper.Keeper
	GenesisState    = types.GenesisState
	Currency        = types.Currency
	CurrencyParams  = types.CurrencyParams
	ResCurrencyInfo = types.ResCurrencyInfo
	ResBalance      = types.ResBalance
	Balance         = types.Balance
	Balances        = types.Balances
)

const (
	ModuleName        = types.ModuleName
	StoreKey          = types.StoreKey
	DefaultParamspace = types.DefaultParamspace
	// Event types, attribute types and values
	EventTypesCreate  = types.EventTypesCreate
	//
	AttributeDenom    = types.AttributeDenom
	AttributeDecimals = types.AttributeDecimals
	AttributeInfoPath = types.AttributeInfoPath
	// Permissions
	PermCCCreator    = types.PermCCCreator
	PermCCUpdater    = types.PermCCUpdater
	PermCCReader     = types.PermCCReader
	PermCCResUpdater = types.PermCCResUpdater
)

var (
	// variable aliases
	ModuleCdc            = types.ModuleCdc
	AvailablePermissions = types.AvailablePermissions
	// function aliases
	NewKeeper           = keeper.NewKeeper
	DefaultGenesisState = types.DefaultGenesisState
	// errors
	ErrInternal    = types.ErrInternal
	ErrWrongDenom  = types.ErrWrongDenom
	ErrWrongParams = types.ErrWrongParams
)
