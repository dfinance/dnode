package currencies_register

import (
	"github.com/dfinance/dnode/x/currencies_register/internal/keeper"
	"github.com/dfinance/dnode/x/currencies_register/internal/types"
)

const (
	ModuleName = types.ModuleName
	StoreKey   = types.StoreKey
)

type (
	Keeper          = keeper.Keeper
	GenesisState    = types.GenesisState
	GenesisCurrency = types.GenesisCurrency
	CurrencyInfo    = types.CurrencyInfo
)

var (
	NewKeeper = keeper.NewKeeper
)
