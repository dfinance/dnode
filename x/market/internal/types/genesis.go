package types

import "fmt"

type GenesisState struct {
	Params Params `json:"params" yaml:"params"`
}

func ValidateGenesis(data GenesisState) error {
	if err := data.Params.Validate(); err != nil {
		return fmt.Errorf("params: %w", err)
	}

	return nil
}

func NewGenesisState(p Params) GenesisState {
	return GenesisState{
		Params: p,
	}
}

func DefaultGenesisState() GenesisState {
	return NewGenesisState(
		DefaultParams(),
	)
}
