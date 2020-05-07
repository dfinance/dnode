package market

import (
	"github.com/dfinance/dnode/x/market/internal/keeper"
	"github.com/dfinance/dnode/x/market/internal/types"
)

type (
	Keeper  = keeper.Keeper
	Market  = types.Market
	Markets = types.Markets
)

const (
	ModuleName        = types.ModuleName
	DefaultParamspace = types.DefaultParamspace
)

var (
	// variable aliases
	ModuleCdc = types.ModuleCdc
	// function aliases
	NewKeeper = keeper.NewKeeper
)
