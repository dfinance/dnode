package orderbook

import (
	"github.com/dfinance/dnode/x/orderbook/internal/keeper"
	"github.com/dfinance/dnode/x/orderbook/internal/types"
)

type (
	Keeper = keeper.Keeper
)

const (
	ModuleName   = types.ModuleName
)

var (
	// variable aliases
	ModuleCdc = types.ModuleCdc
	// function aliases
	NewKeeper = keeper.NewKeeper
)
