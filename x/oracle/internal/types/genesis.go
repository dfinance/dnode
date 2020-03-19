package types

import (
	"bytes"
)

// GenesisState - oracle state that must be provided at genesis
type GenesisState struct {
	Params Params `json:"asset_params" yaml:"asset_params"`
	Assets Assets `json:"assets" yaml:"assets"`
}

// NewGenesisState creates a new genesis state for the oracle module
func NewGenesisState(p Params, assets Assets) GenesisState {
	return GenesisState{
		Params: p,
		Assets: assets,
	}
}

// DefaultGenesisState defines default GenesisState for oracle
func DefaultGenesisState() GenesisState {
	return NewGenesisState(
		DefaultParams(),
		Assets{},
	)
}

// Equal checks whether two gov GenesisState structs are equivalent
func (data GenesisState) Equal(data2 GenesisState) bool {
	b1 := ModuleCdc.MustMarshalBinaryBare(data)
	b2 := ModuleCdc.MustMarshalBinaryBare(data2)
	return bytes.Equal(b1, b2)
}

// IsEmpty returns true if a GenesisState is empty
func (data GenesisState) IsEmpty() bool {
	return data.Equal(GenesisState{})
}

// ValidateGenesis performs basic validation of genesis data returning an
// error for any failed validation criteria.
func ValidateGenesis(data GenesisState) error {
	if err := data.Params.Validate(); err != nil {
		return err
	}
	return nil
}
