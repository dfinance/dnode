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
		err := state.Validate(time.Now())

		require.Error(t, err)
		require.Contains(t, err.Error(), "id")
		require.Contains(t, err.Error(), "nil")
	}

	//validateGenesis wrong owner
	{
		state := getTestGenesisState()
		state.Orders[0].Owner = sdk.AccAddress{}
		err := state.Validate(time.Now())

		require.Error(t, err)
		require.Contains(t, err.Error(), "owner")
		require.Contains(t, err.Error(), "empty")
	}

	// wrong owner
	{
		state := getTestGenesisState()
		state.Orders[0].Owner = sdk.AccAddress(strings.Repeat("0", 50))
		err := state.Validate(time.Now())

		require.Error(t, err)
		require.Contains(t, err.Error(), "owner")
		require.Contains(t, err.Error(), "format")
		require.Contains(t, err.Error(), "wrong")
	}

	// wrong direction
	{
		state := getTestGenesisState()
		state.Orders[0].Direction = "wrong"
		err := state.Validate(time.Now())

		require.Error(t, err)
		require.Contains(t, err.Error(), "direction")
		require.Contains(t, err.Error(), "invalid")
	}

	// wrong price
	{
		state := getTestGenesisState()
		state.Orders[0].Price = sdk.NewUint(0)
		err := state.Validate(time.Now())

		require.Error(t, err)
		require.Contains(t, err.Error(), "price")
		require.Contains(t, err.Error(), "zero")
	}

	// wrong quantity
	{
		state := getTestGenesisState()
		state.Orders[0].Quantity = sdk.NewUint(0)
		err := state.Validate(time.Now())

		require.Error(t, err)
		require.Contains(t, err.Error(), "quantity")
		require.Contains(t, err.Error(), "zero")
	}

	// wrong dates
	{
		state := getTestGenesisState()
		state.Orders[0].CreatedAt = time.Unix(2, 0)
		state.Orders[0].UpdatedAt = time.Unix(1, 0)
		err := state.Validate(time.Now())

		require.Error(t, err)
		require.Contains(t, err.Error(), "wrong create and update dates")
	}

	// future dates
	{
		// CreatedAt
		{
			state := getTestGenesisState()
			state.Orders[0].CreatedAt = time.Now().Add(time.Hour * 10)
			state.Orders[0].UpdatedAt = time.Now().Add(time.Hour * 11)
			err := state.Validate(time.Now())

			require.Error(t, err)
			require.Contains(t, err.Error(), "created_at")
			require.Contains(t, err.Error(), "future date")
		}

		// UpdatedAt
		{
			state := getTestGenesisState()
			state.Orders[0].CreatedAt = time.Now().Truncate(time.Hour)
			state.Orders[0].UpdatedAt = time.Now().Add(time.Hour * 1)
			err := state.Validate(time.Now())

			require.Error(t, err)
			require.Contains(t, err.Error(), "updated_at")
			require.Contains(t, err.Error(), "future date")
		}
	}

	// zero dates
	{
		// CreatedAt
		{
			state := getTestGenesisState()
			state.Orders[0].CreatedAt = time.Time{}
			err := state.Validate(time.Now())

			require.Error(t, err)
			require.Contains(t, err.Error(), "created_at")
			require.Contains(t, err.Error(), "zero")
		}
	}

	// wrong market id
	{
		state := getTestGenesisState()
		state.Orders[0].Market.ID, _ = types.NewIDFromString("")
		err := state.Validate(time.Now())

		require.Error(t, err)
		require.Contains(t, err.Error(), "market")
		require.Contains(t, err.Error(), "id")
		require.Contains(t, err.Error(), "nil")
	}

	// empty market BaseCurrency Denom
	{
		state := getTestGenesisState()
		state.Orders[0].Market.BaseCurrency.Denom = ""
		err := state.Validate(time.Now())

		require.Error(t, err)
		require.Contains(t, err.Error(), "market")
		require.Contains(t, err.Error(), "denom")
		require.Contains(t, err.Error(), "base")
		require.Contains(t, err.Error(), "empty")
	}

	// wrong market BaseCurrency Denom
	{
		state := getTestGenesisState()
		state.Orders[0].Market.BaseCurrency.Denom = "wrong_denom"
		err := state.Validate(time.Now())

		require.Error(t, err)
		require.Contains(t, err.Error(), "market")
		require.Contains(t, err.Error(), "denom")
		require.Contains(t, err.Error(), "base_currency")
		require.Contains(t, err.Error(), "invalid")
	}

	// empty market QuoteCurrency Denom
	{
		state := getTestGenesisState()
		state.Orders[0].Market.QuoteCurrency.Denom = ""
		err := state.Validate(time.Now())

		require.Error(t, err)
		require.Contains(t, err.Error(), "market")
		require.Contains(t, err.Error(), "denom")
		require.Contains(t, err.Error(), "quote")
		require.Contains(t, err.Error(), "empty")
	}

	// wrong market QuoteCurrency Denom
	{
		state := getTestGenesisState()
		state.Orders[0].Market.QuoteCurrency.Denom = "wrong_denom"
		err := state.Validate(time.Now())

		require.Error(t, err)
		require.Contains(t, err.Error(), "market")
		require.Contains(t, err.Error(), "denom")
		require.Contains(t, err.Error(), "quote_currency")
		require.Contains(t, err.Error(), "invalid")
	}

	// empty orders, existiong lastId
	{
		id := types.NewIDFromUint64(1)
		err := GenesisState{LastOrderId: &id}.Validate(time.Now())

		require.Error(t, err)
		require.Contains(t, err.Error(), "last_order_id")
		require.Contains(t, err.Error(), "not nil")
		require.Contains(t, err.Error(), "without")
		require.Contains(t, err.Error(), "orders")
	}

	// empty orders, without lastId
	{
		order := NewMockOrder()
		err := GenesisState{Orders: Orders{order}}.Validate(time.Now())

		require.Error(t, err)
		require.Contains(t, err.Error(), "last_order_id")
		require.Contains(t, err.Error(), "nil")
		require.Contains(t, err.Error(), "existing")
		require.Contains(t, err.Error(), "orders")
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
		}.Validate(time.Now())

		require.Error(t, err)
		require.Contains(t, err.Error(), "last_order_id")
		require.Contains(t, err.Error(), "not equal to max order ID")
	}

	// updatedAt later than block time
	{
		state := getTestGenesisState()
		tmpT := time.Now().Truncate(time.Hour)
		state.Orders[0].UpdatedAt = tmpT
		state.Orders[0].CreatedAt = tmpT
		err := state.Validate(time.Now().Truncate(time.Hour * 2))

		require.Error(t, err)
		require.Contains(t, err.Error(), "create_at")
		require.Contains(t, err.Error(), "after block time")
	}

	// updatedAt later than block time
	{
		state := getTestGenesisState()
		tmpT := time.Now().Truncate(time.Hour)
		state.Orders[0].UpdatedAt = tmpT
		state.Orders[0].CreatedAt = tmpT
		err := state.Validate(time.Time{})

		require.Nil(t, err)
	}
}
