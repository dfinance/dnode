package types

import (
	"bytes"
	"fmt"

	dnTypes "github.com/dfinance/dnode/helpers/types"
)

// GenesisState orders state that must be provided at genesis.
type GenesisState struct {
	Orders      Orders     `json:"orders" yaml:"orders"`
	LastOrderId dnTypes.ID `json:"last_order_id" yaml:"last_order_id"`
}

// Validate checks that genesis state is valid.
func (gs GenesisState) Validate() error {
	for i, item := range gs.Orders {
		if err := item.Valid(); err != nil {
			return fmt.Errorf("order[%d]: %w", i, err)
		}
	}

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

// DefaultGenesisState defines default GenesisState for oracle.
func DefaultGenesisState() GenesisState {
	return GenesisState{
		Orders: Orders{},
	}
}
