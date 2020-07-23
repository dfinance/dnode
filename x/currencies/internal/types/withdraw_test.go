//+build unit

package types

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	dnTypes "github.com/dfinance/dnode/helpers/types"
)

// Withdraw genesis validation.
func TestCurrencies_Withdraw_Valid(t *testing.T) {
	withdraw := Withdraw{}

	// fail: id invalid
	{
		require.Error(t, withdraw.Valid(time.Time{}))
	}
	// fail: coin.denom invalid
	{
		withdraw.ID = dnTypes.NewZeroID()
		withdraw.Coin = sdk.NewCoin("eth1", sdk.OneInt())
		require.Error(t, withdraw.Valid(time.Time{}))
	}
	// fail: coin.amount == 0
	{
		withdraw.Coin = sdk.NewCoin("eth", sdk.ZeroInt())
		require.Error(t, withdraw.Valid(time.Time{}))
	}
	// fail: spender empty
	{
		withdraw.Coin = sdk.NewCoin("eth", sdk.OneInt())
		require.Error(t, withdraw.Valid(time.Time{}))
	}
	// fail: pegZoneSpender empty
	{
		withdraw.Spender = sdk.AccAddress("addr1")
		require.Error(t, withdraw.Valid(time.Time{}))
	}
	// fail: timestamp == 0
	{
		withdraw.PegZoneSpender = "spender"
		require.Error(t, withdraw.Valid(time.Time{}))
	}
	// fail: timestamp is after curTime
	{
		now := time.Now()
		withdraw.Timestamp = now.Add(1 * time.Second).Unix()
		require.Error(t, withdraw.Valid(now))
	}
	// fail: txHash empty
	{
		require.Error(t, withdraw.Valid(time.Time{}))
	}
	// ok
	{
		withdraw.TxHash = "hash"
		require.NoError(t, withdraw.Valid(time.Time{}))
	}
}
