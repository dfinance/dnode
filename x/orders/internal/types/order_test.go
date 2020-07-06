// +build unit

package types

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/cc_storage"
	"github.com/dfinance/dnode/x/markets"
)

func NewMockOrder() Order {
	now := time.Now()

	return Order{
		ID:    dnTypes.NewIDFromUint64(0),
		Owner: sdk.AccAddress("wallet13jyjuz3kkdvqw8u4qfkwd94emdl3vx394kn07h"),
		Market: markets.MarketExtended{
			ID: dnTypes.NewIDFromUint64(0),
			BaseCurrency: cc_storage.Currency{
				Denom:    "btc",
				Decimals: 8,
			},
			QuoteCurrency: cc_storage.Currency{
				Denom:    "dfi",
				Decimals: 18,
			},
		},
		Direction: Bid,
		Price:     sdk.NewUintFromString("1000000000000000000"),
		Quantity:  sdk.NewUintFromString("100000000"),
		Ttl:       60,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func TestOrders_Order_ValidatePriceQuantity(t *testing.T) {
	orderOk := NewMockOrder()

	// ok
	require.NoError(t, orderOk.ValidatePriceQuantity())

	// fail: price
	{
		orderFail := orderOk
		orderFail.Price = sdk.ZeroUint()
		require.Error(t, orderFail.ValidatePriceQuantity())
	}

	// fail: quantity
	{
		orderFail := orderOk
		orderFail.Quantity = sdk.ZeroUint()
		require.Error(t, orderFail.ValidatePriceQuantity())
	}
}

func TestOrders_Order_LockCoin(t *testing.T) {
	order := NewMockOrder()

	// bid order
	{
		// price is too low
		{
			bidOrder := order
			bidOrder.Direction = Bid
			bidOrder.Price = sdk.OneUint()
			bidOrder.Quantity = sdk.OneUint()

			_, err := bidOrder.LockCoin()
			require.Error(t, err)
		}

		// ok
		{
			bidOrder := order
			bidOrder.Direction = Bid

			coin, err := bidOrder.LockCoin()
			require.NoError(t, err)
			require.Equal(t, coin.Denom, string(bidOrder.Market.QuoteCurrency.Denom))
			require.False(t, coin.Amount.IsZero())
		}
	}

	// ask order
	{
		askOrder := order
		askOrder.Direction = Ask

		coin, err := askOrder.LockCoin()
		require.NoError(t, err)
		require.Equal(t, coin.Denom, string(askOrder.Market.BaseCurrency.Denom))
		require.False(t, coin.Amount.IsZero())
	}

	// unsupported type
	{
		failOrder := order
		failOrder.Direction = ""
		_, err := failOrder.LockCoin()
		require.Error(t, err)
	}
}
