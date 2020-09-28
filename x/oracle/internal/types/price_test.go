// +build unit

package types

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	dnTypes "github.com/dfinance/dnode/helpers/types"
)

func NewMockCurrentPrice(assetCode string, ask, bid int64) CurrentPrice {
	return CurrentPrice{
		AssetCode:  dnTypes.AssetCode(assetCode),
		AskPrice:   sdk.NewInt(ask),
		BidPrice:   sdk.NewInt(bid),
		ReceivedAt: time.Now(),
	}
}

func TestOracle_Price_Valid(t *testing.T) {
	// ok
	{
		price := NewMockCurrentPrice("btc_xfi", 101, 100)
		err := price.Valid()
		require.Nil(t, err)
	}

	// wrong asset code
	{
		price := NewMockCurrentPrice("btcXfi", 101, 100)
		err := price.Valid()
		require.Error(t, err)
		require.Contains(t, err.Error(), "asset_code")
	}

	// wrong price: zero
	{
		price := NewMockCurrentPrice("btc_xfi", 0, 0)
		err := price.Valid()
		require.Error(t, err)
		require.Contains(t, err.Error(), "Price")
		require.Contains(t, err.Error(), "is zero")
	}

	// wrong price: negative
	{
		price := NewMockCurrentPrice("btc_xfi", -1, -1)
		err := price.Valid()
		require.Error(t, err)
		require.Contains(t, err.Error(), "Price")
		require.Contains(t, err.Error(), "negative")
	}

	// wrong ReceivedAt: zero
	{
		price := NewMockCurrentPrice("btc_xfi", 100, 99)
		price.ReceivedAt = time.Time{}
		err := price.Valid()
		require.Error(t, err)
		require.Contains(t, err.Error(), "received_at")
		require.Contains(t, err.Error(), "zero")
	}
}

func TestOracle_Price_GetReversedAssetCurrentPrice(t *testing.T) {
	// calculate reverse price ask: 10900.55, bid: 10889.95
	{
		price := NewMockCurrentPrice("btc_xfi", 1090055000000, 1088995000000)
		rp := price.GetReversedAssetCurrentPrice()

		require.Equal(t, rp.AssetCode.String(), "xfi_btc")
		require.Equal(t, rp.AskPrice.String(), "9182") // (1/bid/10^8) * 10^8
		require.Equal(t, rp.BidPrice.String(), "9173") // (1/ask/10^8) * 10^8
	}
}
