// +build unit

package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

// Issue genesis validation.
func TestCurrencies_Issue_Valid(t *testing.T) {
	issue := Issue{}

	// fail: coin.denom invalid
	{
		issue.Coin = sdk.NewCoin("eth1", sdk.OneInt())
		require.Error(t, issue.Valid())
	}
	// fail: coin.amount == 0
	{
		issue.Coin = sdk.NewCoin("eth", sdk.ZeroInt())
		require.Error(t, issue.Valid())
	}
	// fail: payee empty
	{
		issue.Coin = sdk.NewCoin("eth", sdk.NewInt(100))
		require.Error(t, issue.Valid())
	}
	// ok
	{
		issue.Payee = sdk.AccAddress("addr1")
		require.NoError(t, issue.Valid())
	}
}
