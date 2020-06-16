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
				Path:        "0172c9f1bfe0a2bf6ac342aaa3c3380852d4694ae4e71655d37aa5d2e6700ed94e",
				Denom:       "dfi",
				Decimals:    18,
				TotalSupply: dfiVal,
			},
		},
	}
}
