package types

import (
	"fmt"

	dnTypes "github.com/dfinance/dnode/helpers/types"
)

// GenesisState is module's genesis (initial state).
type GenesisState struct {
	CurrenciesParams CurrenciesParams `json:"currencies_params" yaml:"currencies_params"`
}

// Validate checks that genesis state is valid.
func (s GenesisState) Validate() error {
	denomsSet := make(map[string]bool)
	for _, params := range s.CurrenciesParams {
		if denomsSet[params.Denom] {
			return fmt.Errorf("params for %q: duplicated", params.Denom)
		}

		if err := params.Validate(); err != nil {
			return fmt.Errorf("params for %q: %w", params.Denom, err)
		}

		denomsSet[params.Denom] = true
	}

	return nil
}

// CurrencyParams defines currency genesis params and currency cration params.
type CurrencyParams struct {
	// Denomination symbol
	Denom string `json:"denom" yaml:"denom"`
	// Currency decimals count
	Decimals uint8 `json:"decimals" yaml:"decimals"`
}

// Validate check that params are valid.
func (c CurrencyParams) Validate() error {
	if err := dnTypes.DenomFilter(c.Denom); err != nil {
		return fmt.Errorf("denom: %w", err)
	}
	return nil
}

// CurrenciesParams slice of CurrencyParams objects.
type CurrenciesParams []CurrencyParams

// DefaultGenesisState returns default genesis state (validation is done on module init).
func DefaultGenesisState() GenesisState {
	state := GenesisState{
		CurrenciesParams: CurrenciesParams{
			{
				Denom:    "xfi",
				Decimals: 18,
			},
			{
				Denom:    "sxfi",
				Decimals: 18,
			},
			{
				Denom:    "eth",
				Decimals: 18,
			},
			{
				Denom:    "usdt",
				Decimals: 6,
			},
			{
				Denom:    "btc",
				Decimals: 8,
			},
		},
	}

	return state
}
