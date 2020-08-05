// +build unit

package types

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	dnTypes "github.com/dfinance/dnode/helpers/types"
)

func NewMockCurrentPrice(assetCode string, price int64) CurrentPrice {
	return CurrentPrice{
		AssetCode:  dnTypes.AssetCode(assetCode),
		Price:      sdk.NewInt(price),
		ReceivedAt: time.Now(),
	}
}

func TestOracle_Price_Valid(t *testing.T) {
	// ok
	{
		price := NewMockCurrentPrice("btc_dfi", 100)
		err := price.Valid()
		require.Nil(t, err)
	}

	// wrong asset code
	{
		price := NewMockCurrentPrice("btcDfi", 100)
		err := price.Valid()
		require.Error(t, err)
		require.Contains(t, err.Error(), "asset_code")
	}

	// wrong price: zero
	{
		price := NewMockCurrentPrice("btc_dfi", 0)
		err := price.Valid()
		require.Error(t, err)
		require.Contains(t, err.Error(), "price")
		require.Contains(t, err.Error(), "is zero")
	}

	// wrong price: negative
	{
		price := NewMockCurrentPrice("btc_dfi", -1)
		err := price.Valid()
		require.Error(t, err)
		require.Contains(t, err.Error(), "price")
		require.Contains(t, err.Error(), "negative")
	}

	// wrong ReceivedAt: zero
	{
		price := NewMockCurrentPrice("btc_dfi", 100)
		price.ReceivedAt = time.Time{}
		err := price.Valid()
		require.Error(t, err)
		require.Contains(t, err.Error(), "received_at")
		require.Contains(t, err.Error(), "zero")
	}
}
