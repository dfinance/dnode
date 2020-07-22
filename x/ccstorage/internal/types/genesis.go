package types

import (
	"encoding/hex"
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
	// Path used to store account balance for currency denom (0x1::Dfinance::T<Coin>)
	BalancePathHex string `json:"balance_path_hex" yaml:"balance_path_hex"`
	// Path used to store CurrencyInfo for currency denom (0x1::Dfinance::Info<Coin>)
	InfoPathHex string `json:"info_path_hex" yaml:"info_path_hex"`
}

// Validate check that params are valid.
func (c CurrencyParams) Validate() error {
	if err := dnTypes.DenomFilter(c.Denom); err != nil {
		return fmt.Errorf("denom: %w", err)
	}
	if len(c.BalancePathHex) == 0 {
		return fmt.Errorf("balancePathHex: empty")
	}
	if len(c.InfoPathHex) == 0 {
		return fmt.Errorf("infoPathHex: empty")
	}
	if _, err := hex.DecodeString(c.BalancePathHex); err != nil {
		return fmt.Errorf("balancePathHex: %w", err)
	}
	if _, err := hex.DecodeString(c.InfoPathHex); err != nil {
		return fmt.Errorf("infoPathHex: %w", err)
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
				Denom:          "dfi",
				Decimals:       18,
				BalancePathHex: "01608540feb9c6bd277405cfdc0e9140c1431f236f7d97865575e830af3dd67e7e",
				InfoPathHex:    "01f3a1f15d7b13931f3bd5f957ad154b5cbaa0e1a2c3d4d967f286e8800eeb510d",
			},
			{
				Denom:          "eth",
				Decimals:       18,
				BalancePathHex: "0138f4f2895881c804de0e57ced1d44f02e976f9c6561c889f7b7eef8e660d2c9a",
				InfoPathHex:    "012a00668b5325f832c28a24eb83dffa8295170c80345fbfbf99a5263f962c76f4",
			},
			{
				Denom:          "usdt",
				Decimals:       6,
				BalancePathHex: "01a04b6467f35792e0fda5638a509cc807b3b289a4e0ea10794c7db5dc1a63d481",
				InfoPathHex:    "01d058943a984bc02bc4a8547e7c0d780c59334e9aa415b90c87e70d140b2137b8",
			},
			{
				Denom:          "btc",
				Decimals:       8,
				BalancePathHex: "019a2b233aea4cab2e5b6701280f8302be41ea5731af93858fd96e038499eda072",
				InfoPathHex:    "019fdf92aeba5356ec5455b1246c2e1b71d5c7192c6e5a1b50444dafaedc1c40c9",
			},
		},
	}

	return state
}
