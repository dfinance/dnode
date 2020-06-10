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
				Path:        "01d24136b8144bf1669f04b59f88edcb845d9eaf62c2440509c4945f4bc2213494",
				Denom:       "dfi",
				Decimals:    18,
				TotalSupply: dfiVal,
			},
		},
	}
}
