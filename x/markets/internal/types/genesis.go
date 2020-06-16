package types

import "fmt"

// Module genesis state object.
type GenesisState struct {
	Params Params `json:"params" yaml:"params"`
}

// ValidateGenesis validates module genesis state.
func ValidateGenesis(data GenesisState) error {
	if err := data.Params.Validate(); err != nil {
		return fmt.Errorf("params: %w", err)
	}

	return nil
}

// NewGenesisState creates new module genesis state.
func NewGenesisState(p Params) GenesisState {
	return GenesisState{
		Params: p,
	}
}

// DefaultGenesisState returns module default genesis state.
func DefaultGenesisState() GenesisState {
	return NewGenesisState(
		DefaultParams(),
	)
}
