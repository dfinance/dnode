// +build unit

package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/dfinance/dnode/x/orders"
)

type AggInput struct {
	Input  orders.Orders
	Output OrderAggregates
}

func (aggInput AggInput) Check(t *testing.T, agg OrderAggregates) {
	require.Len(t, agg, len(aggInput.Output))
	for i := range agg {
		require.True(t, aggInput.Output[i].Price.Equal(agg[i].Price), "%d: Price (expected / received): %s / %s", i, aggInput.Output[i].Price, agg[i].Price)
		require.True(t, aggInput.Output[i].Quantity.Equal(agg[i].Quantity), "%d: Quantity (expected / received): %s / %s", i, aggInput.Output[i].Quantity, agg[i].Quantity)
	}
}

func TestOBKeeper_OrderAggregate_Bid(t *testing.T) {
	t.Parallel()

	// zero input
	{
		agg := NewBidOrderAggregates(orders.Orders{})
		require.Len(t, agg, 0)
	}

	// one input
	{
		input := AggInput{
			Input: orders.Orders{
				orders.Order{Price: sdk.NewUint(50), Quantity: sdk.NewUint(100)},
			},
			Output: OrderAggregates{
				OrderAggregate{Price: sdk.NewUint(50), Quantity: sdk.NewUint(100)},
			},
		}
		input.Check(t, NewBidOrderAggregates(input.Input))
	}

	// norm input
	{
		input := AggInput{
			Input: orders.Orders{
				orders.Order{Price: sdk.NewUint(10), Quantity: sdk.NewUint(50)},
				orders.Order{Price: sdk.NewUint(10), Quantity: sdk.NewUint(50)},
				orders.Order{Price: sdk.NewUint(50), Quantity: sdk.NewUint(100)},
				orders.Order{Price: sdk.NewUint(50), Quantity: sdk.NewUint(150)},
				orders.Order{Price: sdk.NewUint(50), Quantity: sdk.NewUint(50)},
				orders.Order{Price: sdk.NewUint(75), Quantity: sdk.NewUint(150)},
				orders.Order{Price: sdk.NewUint(150), Quantity: sdk.NewUint(200)},
				orders.Order{Price: sdk.NewUint(150), Quantity: sdk.NewUint(100)},
			},
			Output: OrderAggregates{
				OrderAggregate{Price: sdk.NewUint(10), Quantity: sdk.NewUint(850)},
				OrderAggregate{Price: sdk.NewUint(50), Quantity: sdk.NewUint(750)},
				OrderAggregate{Price: sdk.NewUint(75), Quantity: sdk.NewUint(450)},
				OrderAggregate{Price: sdk.NewUint(150), Quantity: sdk.NewUint(300)},
			},
		}
		input.Check(t, NewBidOrderAggregates(input.Input))
	}
}

func TestOBKeeper_OrderAggregate_Ask(t *testing.T) {
	t.Parallel()

	// zero input
	{
		agg := NewAskOrderAggregates(orders.Orders{})
		require.Len(t, agg, 0)
	}

	// one input
	{
		input := AggInput{
			Input: orders.Orders{
				orders.Order{Price: sdk.NewUint(50), Quantity: sdk.NewUint(100)},
			},
			Output: OrderAggregates{
				OrderAggregate{Price: sdk.NewUint(50), Quantity: sdk.NewUint(100)},
			},
		}
		input.Check(t, NewAskOrderAggregates(input.Input))
	}

	// norm input
	{
		input := AggInput{
			Input: orders.Orders{
				orders.Order{Price: sdk.NewUint(10), Quantity: sdk.NewUint(50)},
				orders.Order{Price: sdk.NewUint(10), Quantity: sdk.NewUint(50)},
				orders.Order{Price: sdk.NewUint(50), Quantity: sdk.NewUint(100)},
				orders.Order{Price: sdk.NewUint(50), Quantity: sdk.NewUint(150)},
				orders.Order{Price: sdk.NewUint(50), Quantity: sdk.NewUint(50)},
				orders.Order{Price: sdk.NewUint(75), Quantity: sdk.NewUint(150)},
				orders.Order{Price: sdk.NewUint(150), Quantity: sdk.NewUint(200)},
				orders.Order{Price: sdk.NewUint(150), Quantity: sdk.NewUint(100)},
			},
			Output: OrderAggregates{
				OrderAggregate{Price: sdk.NewUint(10), Quantity: sdk.NewUint(100)},
				OrderAggregate{Price: sdk.NewUint(50), Quantity: sdk.NewUint(400)},
				OrderAggregate{Price: sdk.NewUint(75), Quantity: sdk.NewUint(550)},
				OrderAggregate{Price: sdk.NewUint(150), Quantity: sdk.NewUint(850)},
			},
		}
		input.Check(t, NewAskOrderAggregates(input.Input))
	}
}
