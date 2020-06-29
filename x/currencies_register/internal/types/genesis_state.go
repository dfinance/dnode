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
				Path:        "01f3a1f15d7b13931f3bd5f957ad154b5cbaa0e1a2c3d4d967f286e8800eeb510d",
				Denom:       "dfi",
				Decimals:    18,
				TotalSupply: dfiVal,
			},
		},
	}
}
