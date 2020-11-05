package ccstorage

import (
	"github.com/dfinance/dnode/x/ccstorage/internal/keeper"
	"github.com/dfinance/dnode/x/ccstorage/internal/types"
)

type (
	Keeper          = keeper.Keeper
	GenesisState    = types.GenesisState
	Currency        = types.Currency
	Currencies      = types.Currencies
	CurrencyParams  = types.CurrencyParams
	ResCurrencyInfo = types.ResCurrencyInfo
	ResBalance      = types.ResBalance
	Balance         = types.Balance
	Balances        = types.Balances
	//
	SquashOptions = keeper.SquashOptions
)

const (
	ModuleName = types.ModuleName
	StoreKey   = types.StoreKey
	// Event types, attribute types and values
	EventTypesCreate = types.EventTypesCreate
	//
	AttributeDenom    = types.AttributeDenom
	AttributeDecimals = types.AttributeDecimals
	AttributeInfoPath = types.AttributeInfoPath
)

var (
	// variable aliases
	ModuleCdc            = types.ModuleCdc
	AvailablePermissions = types.AvailablePermissions
	// function aliases
	NewKeeper           = keeper.NewKeeper
	DefaultGenesisState = types.DefaultGenesisState
	//
	NewEmptySquashOptions = keeper.NewEmptySquashOptions
	// perms requests
	RequestVMStoragePerms = types.RequestVMStoragePerms
	// errors
	ErrInternal    = types.ErrInternal
	ErrWrongDenom  = types.ErrWrongDenom
	ErrWrongParams = types.ErrWrongParams
)
