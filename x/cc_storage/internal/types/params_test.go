// +build unit

package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Test keeper CreateCurrency method.
func TestCCS_CurrencyParams_Validate(t *testing.T) {
	t.Parallel()

	// ok
	{
		param := CurrencyParams{0, "0102", "AABB"}
		require.NoError(t, param.Validate())
	}

	// fail: empty path
	{
		param1 := CurrencyParams{0, "", "AABB"}
		require.Error(t, param1.Validate())

		param2 := CurrencyParams{0, "0102", ""}
		require.Error(t, param2.Validate())
	}

	// fail: invalid hex path
	{
		param1 := CurrencyParams{0, "z", "AABB"}
		require.Error(t, param1.Validate())

		param2 := CurrencyParams{0, "0102", "z"}
		require.Error(t, param2.Validate())
	}
}
