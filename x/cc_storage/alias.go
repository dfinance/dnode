package cc_storage

import (
	"github.com/dfinance/dnode/x/cc_storage/internal/keeper"
	"github.com/dfinance/dnode/x/cc_storage/internal/types"
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
)

var (
	// variable aliases
	ModuleCdc = types.ModuleCdc
	// function aliases
	NewKeeper           = keeper.NewKeeper
	DefaultGenesisState = types.DefaultGenesisState
	// errors
	ErrInternal    = types.ErrInternal
	ErrWrongDenom  = types.ErrWrongDenom
	ErrWrongParams = types.ErrWrongParams
)
