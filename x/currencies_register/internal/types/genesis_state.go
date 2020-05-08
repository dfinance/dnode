package types

import sdk "github.com/cosmos/cosmos-sdk/types"

type GenesisCurrency struct {
	Path        string  `json:"path"`
	Denom       string  `json:"denom"`
	Decimals    uint8   `json:"decimals"`
	TotalSupply sdk.Int `json:"totalSupply"`
}

// Genesis state to add before start.
type GenesisState struct {
	Currencies []GenesisCurrency `json:"currencies"`
}

// Default genesis state with DFI info.
func DefaultGenesisState() GenesisState {
	dfiVal, _ := sdk.NewIntFromString("100000000000000000000000000")

	return GenesisState{
		Currencies: []GenesisCurrency{
			{
				Path:        "018bfc024222e94fbed60ff0c9c1cf48c5b2809d83c82f513b2c385e21ba8a2d35",
				Denom:       "dfi",
				Decimals:    18,
				TotalSupply: dfiVal,
			},
		},
	}
}
