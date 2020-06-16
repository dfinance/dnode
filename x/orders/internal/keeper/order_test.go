// +build unit

package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/orders/internal/types"
)

func Test_Order_StoreIO(t *testing.T) {
	input := NewTestInput(t)

	// check non-existing
	{
		id := dnTypes.NewIDFromUint64(0)
		require.False(t, input.keeper.Has(input.ctx, id))

		_, err := input.keeper.Get(input.ctx, id)
		require.Error(t, err)
	}

	// add order
	inOrder := NewBtcDfiMockOrder(types.Bid)
	{
		input.keeper.Set(input.ctx, inOrder)

		outOrder, err := input.keeper.Get(input.ctx, inOrder.ID)
		require.NoError(t, err)
		CompareOrders(t, inOrder, outOrder)
	}

	// del order
	{
		input.keeper.Del(input.ctx, inOrder.ID)
		require.False(t, input.keeper.Has(input.ctx, inOrder.ID))
	}

	// del deleted
	{
		input.keeper.Del(input.ctx, inOrder.ID)
	}
}

func Test_Order_List(t *testing.T) {
	input := NewTestInput(t)

	// get empty list
	{
		outOrders, err := input.keeper.GetList(input.ctx)
		require.NoError(t, err)
		require.Len(t, outOrders, 0)
	}

	order1 := NewBtcDfiMockOrder(types.Ask)
	order1.ID = dnTypes.NewIDFromUint64(0)
	order1.Price = order1.Price.AddUint64(1000)
	order1.Quantity = order1.Quantity.AddUint64(1000)
	input.keeper.Set(input.ctx, order1)

	order2 := NewEthDfiMockOrder(types.Bid)
	order2.ID = dnTypes.NewIDFromUint64(1)
	order2.Price = order2.Price.AddUint64(1000)
	order2.Quantity = order2.Quantity.AddUint64(1000)
	input.keeper.Set(input.ctx, order2)

	order3 := NewBtcDfiMockOrder(types.Bid)
	order3.ID = dnTypes.NewIDFromUint64(2)
	order3.Price = order3.Price.SubUint64(1000)
	order3.Quantity = order3.Quantity.SubUint64(1000)
	input.keeper.Set(input.ctx, order3)

	order4 := NewEthDfiMockOrder(types.Ask)
	order4.ID = dnTypes.NewIDFromUint64(3)
	order4.Price = order4.Price.SubUint64(1000)
	order4.Quantity = order4.Quantity.SubUint64(1000)
	input.keeper.Set(input.ctx, order4)

	inOrders := types.Orders{order1, order2, order3, order4}

	// check list all
	{
		outOrders, err := input.keeper.GetList(input.ctx)
		require.NoError(t, err)

		require.Len(t, outOrders, len(inOrders))
		for i := range outOrders {
			CompareOrders(t, inOrders[i], outOrders[i])
		}
	}

	// check direct iterator
	{
		iterator := input.keeper.GetIterator(input.ctx)
		defer iterator.Close()

		i := uint64(0)
		for ; iterator.Valid(); iterator.Next() {
			order := types.Order{}
			require.NoError(t, input.cdc.UnmarshalBinaryLengthPrefixed(iterator.Value(), &order))
			require.Equal(t, order.ID.UInt64(), i)
			i++
		}
	}

	// check reverse iterator
	{
		iterator := input.keeper.GetReverseIterator(input.ctx)
		defer iterator.Close()

		i := uint64(len(inOrders) - 1)
		for ; iterator.Valid(); iterator.Next() {
			order := types.Order{}
			require.NoError(t, input.cdc.UnmarshalBinaryLengthPrefixed(iterator.Value(), &order))
			require.Equal(t, order.ID.UInt64(), i)
			i--
		}
	}

	// check filtered list
	{
		// check limit
		{
			params := types.OrdersReq{
				Page:  1,
				Limit: 1,
			}

			outOrders, err := input.keeper.GetListFiltered(input.ctx, params)
			require.NoError(t, err)
			require.Len(t, outOrders, 1)
		}

		// owner filter
		{
			params := types.OrdersReq{
				Page:  1,
				Owner: sdk.AccAddress("wallet13jyjuz3kkdvqw8u4qfkwd94emdl3vx394kn07i"),
			}

			outOrders, err := input.keeper.GetListFiltered(input.ctx, params)
			require.NoError(t, err)
			require.Len(t, outOrders, 2)
			require.Equal(t, outOrders[0].ID.UInt64(), uint64(1))
			require.Equal(t, outOrders[1].ID.UInt64(), uint64(3))
		}

		// direction filter
		{
			params := types.OrdersReq{
				Page:  1,
				Direction: types.Bid,
			}

			outOrders, err := input.keeper.GetListFiltered(input.ctx, params)
			require.NoError(t, err)
			require.Len(t, outOrders, 2)
			require.Equal(t, outOrders[0].ID.UInt64(), uint64(1))
			require.Equal(t, outOrders[1].ID.UInt64(), uint64(2))
		}

		// marketID filter
		{
			params := types.OrdersReq{
				Page:  1,
				MarketID: "0",
			}

			outOrders, err := input.keeper.GetListFiltered(input.ctx, params)
			require.NoError(t, err)
			require.Len(t, outOrders, 2)
			require.Equal(t, outOrders[0].ID.UInt64(), uint64(0))
			require.Equal(t, outOrders[1].ID.UInt64(), uint64(2))
		}

		// check no match
		{
			params := types.OrdersReq{
				Page:  1,
				MarketID: "2",
			}

			outOrders, err := input.keeper.GetListFiltered(input.ctx, params)
			require.NoError(t, err)
			require.Len(t, outOrders, 0)
		}
	}
}
