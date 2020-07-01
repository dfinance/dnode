package types

import "fmt"

// GenesisState is module's genesis (initial state).
type GenesisState struct {
	CurrenciesParams CurrenciesParams `json:"currencies_params"`
}

// Validate checks that genesis state is valid.
func (s GenesisState) Validate() error {
	for denom, params := range s.CurrenciesParams {
		if err := params.Validate(); err != nil {
			return fmt.Errorf("params for %q: %w", denom, err)
		}
	}

	return nil
}

// DefaultGenesisState returns default genesis state (validation is done on module init).
func DefaultGenesisState() GenesisState {
	state := GenesisState{
		CurrenciesParams: make(CurrenciesParams, 0),
	}
	state.CurrenciesParams["dfi"] = CurrencyParams{
		18,
		"01608540feb9c6bd277405cfdc0e9140c1431f236f7d97865575e830af3dd67e7e",
		"01f3a1f15d7b13931f3bd5f957ad154b5cbaa0e1a2c3d4d967f286e8800eeb510d",
	}
	state.CurrenciesParams["eth"] = CurrencyParams{
		18,
		"0138f4f2895881c804de0e57ced1d44f02e976f9c6561c889f7b7eef8e660d2c9a",
		"012a00668b5325f832c28a24eb83dffa8295170c80345fbfbf99a5263f962c76f4",
	}
	state.CurrenciesParams["usdt"] = CurrencyParams{
		6,
		"01a04b6467f35792e0fda5638a509cc807b3b289a4e0ea10794c7db5dc1a63d481",
		"01d058943a984bc02bc4a8547e7c0d780c59334e9aa415b90c87e70d140b2137b8",
	}
	state.CurrenciesParams["btc"] = CurrencyParams{
		8,
		"019a2b233aea4cab2e5b6701280f8302be41ea5731af93858fd96e038499eda072",
		"019fdf92aeba5356ec5455b1246c2e1b71d5c7192c6e5a1b50444dafaedc1c40c9",
	}

	return state
}
