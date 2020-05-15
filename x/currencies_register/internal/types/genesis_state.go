package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// Genesis state for currencies register.
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
				Path:        "011c53cd211c8dd6f27b977dbcf497d6650944f764d15cebf75dcc17f8e2bfa5f4",
				Denom:       "dfi",
				Decimals:    18,
				TotalSupply: dfiVal,
			},
		},
	}
}
