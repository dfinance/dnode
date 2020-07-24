// +build unit

package types

import (
	"strings"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/dfinance/dnode/helpers/types"
)

func getTestGenesisState() GenesisState {
	order := NewMockOrder()
	return GenesisState{
		Orders:      Orders{order},
		LastOrderId: &order.ID,
	}
}

func TestOrders_Genesis_Valid(t *testing.T) {
	//validateGenesis ok
	{
		order := NewMockOrder()
		order.ID = types.NewIDFromUint64(0)
		order2 := NewMockOrder()
		order2.ID = types.NewIDFromUint64(1)
		order2.Market.QuoteCurrency.Denom = "btc"

		order3 := NewMockOrder()
		order3.ID = types.NewIDFromUint64(2)
		orderT := &order3
		order4 := *orderT

		state := GenesisState{
			Orders:      Orders{order, order2, order3},
			LastOrderId: &order3.ID,
		}
		require.NoError(t, state.Validate(time.Now()))
		require.False(t, state.IsEmpty())

		require.False(t, GenesisState{Orders: Orders{order2}}.Equal(GenesisState{Orders: Orders{order3}}))
		require.True(t, GenesisState{Orders: Orders{order3}}.Equal(GenesisState{Orders: Orders{order4}}))
	}

	// wrong id
	{
		state := getTestGenesisState()
		state.Orders[0].ID, _ = types.NewIDFromString("")
		err := state.Validate(time.Now()).Error()
		require.Contains(t, err, "id")
		require.Contains(t, err, "nil")
	}

	//validateGenesis wrong owner
	{
		state := getTestGenesisState()
		state.Orders[0].Owner = sdk.AccAddress{}
		err := state.Validate(time.Now()).Error()
		require.Contains(t, err, "owner")
		require.Contains(t, err, "empty")
	}

	// wrong owner
	{
		state := getTestGenesisState()
		state.Orders[0].Owner = sdk.AccAddress(strings.Repeat("0", 50))
		err := state.Validate(time.Now()).Error()
		require.Contains(t, err, "owner")
		require.Contains(t, err, "format")
		require.Contains(t, err, "wrong")
	}

	// wrong direction
	{
		state := getTestGenesisState()
		state.Orders[0].Direction = "wrong"
		err := state.Validate(time.Now()).Error()
		require.Contains(t, err, "direction")
		require.Contains(t, err, "invalid")
	}

	// wrong price
	{
		state := getTestGenesisState()
		state.Orders[0].Price = sdk.NewUint(0)
		err := state.Validate(time.Now()).Error()
		require.Contains(t, err, "price")
		require.Contains(t, err, "zero")
	}

	// wrong quantity
	{
		state := getTestGenesisState()
		state.Orders[0].Quantity = sdk.NewUint(0)
		err := state.Validate(time.Now()).Error()
		require.Contains(t, err, "quantity")
		require.Contains(t, err, "zero")
	}

	// wrong dates
	{
		state := getTestGenesisState()
		state.Orders[0].CreatedAt = time.Unix(2, 0)
		state.Orders[0].UpdatedAt = time.Unix(1, 0)
		err := state.Validate(time.Now()).Error()
		require.Contains(t, err, "wrong create and update dates")
	}

	// future dates
	{
		// CreatedAt
		{
			state := getTestGenesisState()
			state.Orders[0].CreatedAt = time.Now().Add(time.Hour * 10)
			state.Orders[0].UpdatedAt = time.Now().Add(time.Hour * 11)
			err := state.Validate(time.Now()).Error()
			require.Contains(t, err, "created_at")
			require.Contains(t, err, "future date")
		}

		// UpdatedAt
		{
			state := getTestGenesisState()
			state.Orders[0].CreatedAt = time.Now().Truncate(time.Hour)
			state.Orders[0].UpdatedAt = time.Now().Add(time.Hour * 1)
			err := state.Validate(time.Now()).Error()
			require.Contains(t, err, "updated_at")
			require.Contains(t, err, "future date")
		}
	}

	// zero dates
	{
		// CreatedAt
		{
			state := getTestGenesisState()
			state.Orders[0].CreatedAt = time.Time{}
			err := state.Validate(time.Now()).Error()
			require.Contains(t, err, "created_at")
			require.Contains(t, err, "zero")
		}
	}

	// wrong market id
	{
		state := getTestGenesisState()
		state.Orders[0].Market.ID, _ = types.NewIDFromString("")
		err := state.Validate(time.Now()).Error()
		require.Contains(t, err, "market")
		require.Contains(t, err, "id")
		require.Contains(t, err, "nil")
	}

	// empty market BaseCurrency Denom
	{
		state := getTestGenesisState()
		state.Orders[0].Market.BaseCurrency.Denom = ""
		err := state.Validate(time.Now()).Error()
		require.Contains(t, err, "market")
		require.Contains(t, err, "denom")
		require.Contains(t, err, "base")
		require.Contains(t, err, "empty")
	}

	// wrong market BaseCurrency Denom
	{
		state := getTestGenesisState()
		state.Orders[0].Market.BaseCurrency.Denom = "wrong_denom"
		err := state.Validate(time.Now()).Error()
		require.Contains(t, err, "market")
		require.Contains(t, err, "denom")
		require.Contains(t, err, "base_currency")
		require.Contains(t, err, "invalid")
	}

	// empty market QuoteCurrency Denom
	{
		state := getTestGenesisState()
		state.Orders[0].Market.QuoteCurrency.Denom = ""
		err := state.Validate(time.Now()).Error()
		require.Contains(t, err, "market")
		require.Contains(t, err, "denom")
		require.Contains(t, err, "quote")
		require.Contains(t, err, "empty")
	}

	// wrong market QuoteCurrency Denom
	{
		state := getTestGenesisState()
		state.Orders[0].Market.QuoteCurrency.Denom = "wrong_denom"
		err := state.Validate(time.Now()).Error()
		require.Contains(t, err, "market")
		require.Contains(t, err, "denom")
		require.Contains(t, err, "quote_currency")
		require.Contains(t, err, "invalid")
	}

	// empty orders, existiong lastId
	{
		id := types.NewIDFromUint64(1)
		err := GenesisState{LastOrderId: &id}.Validate(time.Now()).Error()
		require.Contains(t, err, "last_order_id")
		require.Contains(t, err, "not nil")
		require.Contains(t, err, "without")
		require.Contains(t, err, "orders")
	}

	// empty orders, without lastId
	{
		order := NewMockOrder()
		err := GenesisState{Orders: Orders{order}}.Validate(time.Now()).Error()
		require.Contains(t, err, "last_order_id")
		require.Contains(t, err, "nil")
		require.Contains(t, err, "existing")
		require.Contains(t, err, "orders")
	}

	// with orders, wrong lastId
	{
		order := NewMockOrder()
		order.ID = types.NewIDFromUint64(0)
		order2 := NewMockOrder()
		order2.ID = types.NewIDFromUint64(1)

		id := types.NewIDFromUint64(2)

		err := GenesisState{
			Orders:      Orders{order, order2},
			LastOrderId: &id,
		}.Validate(time.Now()).Error()

		require.Contains(t, err, "last_order_id")
		require.Contains(t, err, "not equal to max order ID")
	}

	// updatedAt later than block time
	{
		state := getTestGenesisState()
		state.Orders[0].UpdatedAt = time.Now().Add(time.Hour)
		err := state.Validate(time.Now()).Error()

		require.Contains(t, err, "updated_at")
		require.Contains(t, err, "future date")
	}

	// createdAt later than block time
	{
		state := getTestGenesisState()
		state.Orders[0].CreatedAt = time.Now().Add(time.Hour)
		state.Orders[0].UpdatedAt = time.Now().Add(time.Hour * 2)
		err := state.Validate(time.Now()).Error()

		require.Contains(t, err, "created_at")
		require.Contains(t, err, "future date")
	}
}
