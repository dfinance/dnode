package types

import (
	"bytes"
	"fmt"
	"time"

	dnTypes "github.com/dfinance/dnode/helpers/types"
)

// GenesisState orders state that must be provided at genesis.
type GenesisState struct {
	Orders      Orders      `json:"orders" yaml:"orders"`
	LastOrderId *dnTypes.ID `json:"last_order_id" yaml:"last_order_id"`
}

// Validate checks that genesis state is valid.
func (gs GenesisState) Validate(blockTime time.Time) error {
	maxOrderID := dnTypes.NewZeroID()
	ordersIdsSet := make(map[string]bool, len(gs.Orders))

	for i, order := range gs.Orders {
		if err := order.Valid(); err != nil {
			return fmt.Errorf("order[%d]: %w", i, err)
		}

		if !blockTime.IsZero() && order.CreatedAt.After(blockTime) {
			return fmt.Errorf("order[%d]: create_at after block time", i)
		}

		if !blockTime.IsZero() && order.UpdatedAt.After(blockTime) {
			return fmt.Errorf("order[%d]: updated_at after block time", i)
		}

		if ordersIdsSet[order.ID.String()] {
			return fmt.Errorf("order[%d]: duplicated ID %q", i, order.ID.String())
		}

		ordersIdsSet[order.ID.String()] = true

		if order.ID.GT(maxOrderID) {
			maxOrderID = order.ID
		}
	}

	if gs.LastOrderId == nil && len(gs.Orders) != 0 {
		return fmt.Errorf("last_order_id: nil with existing orders")
	}
	if gs.LastOrderId != nil && len(gs.Orders) == 0 {
		return fmt.Errorf("last_order_id: not nil without existing orders")
	}
	if gs.LastOrderId != nil {
		if err := gs.LastOrderId.Valid(); err != nil {
			return fmt.Errorf("last_order_id: %w", err)
		}

		if !gs.LastOrderId.Equal(maxOrderID) {
			return fmt.Errorf("last_order_id: not equal to max order ID")
		}
	}

	return nil
}

// Equal checks whether two GenesisState structs are equivalent.
func (gs GenesisState) Equal(data2 GenesisState) bool {
	b1 := ModuleCdc.MustMarshalBinaryBare(gs)
	b2 := ModuleCdc.MustMarshalBinaryBare(data2)
	return bytes.Equal(b1, b2)
}

// IsEmpty returns true if a GenesisState is empty.
func (gs GenesisState) IsEmpty() bool {
	return gs.Equal(GenesisState{})
}

// DefaultGenesisState defines default GenesisState for orders.
func DefaultGenesisState() GenesisState {
	return GenesisState{
		Orders: Orders{},
	}
}
