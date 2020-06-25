package currencies_register

import (
	"github.com/dfinance/dnode/x/currencies_register/internal/keeper"
	"github.com/dfinance/dnode/x/currencies_register/internal/types"
)

const (
	ModuleName   = types.ModuleName
	StoreKey     = types.StoreKey
	RouterKey    = types.RouterKey
	GovRouterKey = types.GovRouterKey
)

type (
	Keeper = keeper.Keeper

	GenesisState    = types.GenesisState
	GenesisCurrency = types.GenesisCurrency

	CurrencyInfo = types.CurrencyInfo

	AddCurrencyProposal = types.AddCurrencyProposal
)

var (
	// variable aliases
	ModuleCdc = types.ModuleCdc
	// function aliases
	NewKeeper              = keeper.NewKeeper
	NewQuerier             = keeper.NewQuerier
	RegisterCodec          = types.RegisterCodec
	DefaultGenesisState    = types.DefaultGenesisState
	NewAddCurrencyProposal = types.NewAddCurrencyProposal
	// errors
	ErrGovInvalidProposal = types.ErrGovInvalidProposal
)
