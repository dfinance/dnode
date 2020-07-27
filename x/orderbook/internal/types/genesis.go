package types

import (
	"bytes"
	"time"
)

// GenesisState orderbook state that must be provided at genesis.
type GenesisState struct {
}

// Validate checks that genesis state is valid.
func (gs GenesisState) Validate(blockTime time.Time) error {
	return nil
}

// Equal checks whether two gov GenesisState structs are equivalent.
func (gs GenesisState) Equal(data2 GenesisState) bool {
	b1 := ModuleCdc.MustMarshalBinaryBare(gs)
	b2 := ModuleCdc.MustMarshalBinaryBare(data2)
	return bytes.Equal(b1, b2)
}

// IsEmpty returns true if a GenesisState is empty.
func (gs GenesisState) IsEmpty() bool {
	return gs.Equal(GenesisState{})
}

// DefaultGenesisState defines default GenesisState for orderbook.
func DefaultGenesisState() GenesisState {
	return GenesisState{}
}
