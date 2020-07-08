// +build unit

package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func newMockOrderFill() OrderFill {
	return OrderFill{
		Order:            NewMockOrder(),
		ClearancePrice:   sdk.NewUintFromString("500000000000000000"),
		QuantityFilled:   sdk.NewUintFromString("50000000"),
		QuantityUnfilled: sdk.NewUintFromString("50000000"),
	}
}

func TestOrders_OrderFill_FillCoin(t *testing.T) {
	fill := newMockOrderFill()

	// bid order
	{
		bidFill := fill
		bidFill.Order.Direction = Bid

		coin, err := bidFill.FillCoin()
		require.NoError(t, err)
		require.Equal(t, coin.Denom, string(bidFill.Order.Market.BaseCurrency.Denom))
		require.False(t, coin.Amount.IsZero())
	}

	// ask order
	{
		// ok
		{
			askFill := fill
			askFill.Order.Direction = Ask

			coin, err := askFill.FillCoin()
			require.NoError(t, err)
			require.Equal(t, coin.Denom, string(askFill.Order.Market.QuoteCurrency.Denom))
			require.False(t, coin.Amount.IsZero())
		}

		// price is too low
		{
			askFill := fill
			askFill.Order.Direction = Ask
			askFill.ClearancePrice = sdk.OneUint()

			_, err := askFill.FillCoin()
			require.Error(t, err)
		}
	}

	// unsupported type
	{
		failFill := fill
		failFill.Order.Direction = ""
		_, err := failFill.FillCoin()
		require.Error(t, err)
	}
}

func TestOrders_OrderFill_RefundCoin(t *testing.T) {
	fill := newMockOrderFill()

	// bid order
	{
		// ok
		{
			bidFill := fill
			bidFill.Order.Direction = Bid

			doRefund, coin, err := bidFill.RefundCoin()
			require.NoError(t, err)
			require.True(t, doRefund)
			require.NotNil(t, coin)
			require.Equal(t, coin.Denom, string(bidFill.Order.Market.QuoteCurrency.Denom))
			require.False(t, coin.Amount.IsZero())
		}

		// refund is too small
		{
			bidFill := fill
			bidFill.Order.Direction = Bid
			bidFill.ClearancePrice = bidFill.Order.Price.Sub(sdk.OneUint())

			doRefund, coin, err := bidFill.RefundCoin()
			require.NoError(t, err)
			require.True(t, doRefund)
			require.Nil(t, coin)
		}
	}

	// ask order
	{
		askFill := fill
		askFill.Order.Direction = Ask

		doRefund, coin, err := askFill.RefundCoin()
		require.NoError(t, err)
		require.False(t, doRefund)
		require.Nil(t, coin)
	}

	// unsupported type
	{
		failFill := fill
		failFill.Order.Direction = ""
		doRefund, coin, err := failFill.RefundCoin()
		require.Error(t, err)
		require.False(t, doRefund)
		require.Nil(t, coin)
	}
}
