// +build unit

package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestCCS_Currency_Valid(t *testing.T) {
	t.Parallel()

	// OK
	{
		err := Currency{
			Decimals: 2,
			Denom:    "btc",
			Supply:   sdk.NewInt(100),
		}.Valid()

		require.Nil(t, err)
	}

	// Empty denom
	{
		err := Currency{
			Decimals: 2,
			Supply:   sdk.NewInt(100),
		}.Valid()

		require.Error(t, err)
		require.Contains(t, err.Error(), "denom")
		require.Contains(t, err.Error(), "empty")
	}

	// Wrong denom
	{
		err := Currency{
			Decimals: 2,
			Denom:    "wrong_denom",
			Supply:   sdk.NewInt(100),
		}.Valid()

		require.Error(t, err)
		require.Contains(t, err.Error(), "denom")
		require.Contains(t, err.Error(), "invalid")
	}
}
