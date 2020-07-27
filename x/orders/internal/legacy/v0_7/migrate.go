package v0_7

import (
	"fmt"

	v06 "github.com/dfinance/dnode/x/orders/internal/legacy/v0_6"
)

// Migrate migrates v0.6 module state to v0.7 version.
// - Order.Memo field added;
func Migrate(oldState v06.GenesisState) (GenesisState, error) {
	newState := GenesisState{
		LastOrderId: oldState.LastOrderId,
		Orders:      make(Orders, 0, len(oldState.Orders)),
	}

	for _, oldOrder := range oldState.Orders {
		newOrder := Order{
			ID:        oldOrder.ID,
			Owner:     oldOrder.Owner,
			Market:    oldOrder.Market,
			Direction: oldOrder.Direction,
			Price:     oldOrder.Price,
			Quantity:  oldOrder.Quantity,
			Ttl:       oldOrder.Ttl,
			CreatedAt: oldOrder.CreatedAt,
			UpdatedAt: oldOrder.UpdatedAt,
		}
		newOrder.Memo = fmt.Sprintf("migrated: %s from %s", newOrder.Direction, newOrder.Owner.String())

		newState.Orders = append(newState.Orders, newOrder)
	}

	return newState, nil
}
