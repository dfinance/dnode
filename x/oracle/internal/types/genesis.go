package types

import (
	"bytes"
)

// GenesisState - oracle state that must be provided at genesis
type GenesisState struct {
	Params Params `json:"asset_params" yaml:"asset_params"`
}

// Validate checks that genesis state is valid.
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}

	return nil
}

// DefaultGenesisState defines default GenesisState for oracle
func DefaultGenesisState() GenesisState {
	return GenesisState{
		Params: DefaultParams(),
	}
}

// Equal checks whether two gov GenesisState structs are equivalent
func (gs GenesisState) Equal(data2 GenesisState) bool {
	b1 := ModuleCdc.MustMarshalBinaryBare(gs)
	b2 := ModuleCdc.MustMarshalBinaryBare(data2)
	return bytes.Equal(b1, b2)
}

// IsEmpty returns true if a GenesisState is empty
func (gs GenesisState) IsEmpty() bool {
	return gs.Equal(GenesisState{})
}
