// +build unit

package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMS_Params_Valid(t *testing.T) {
	t.Parallel()

	// ok
	{
		params := Params{MinValidators: DefaultMinValidators, MaxValidators: DefaultMaxValidators}
		require.NoError(t, params.Validate())
	}

	// fail: min
	{
		params := Params{MinValidators: DefaultMinValidators - 1, MaxValidators: DefaultMaxValidators}
		require.Error(t, params.Validate())
	}

	// fail: max
	{
		params := Params{MinValidators: DefaultMinValidators, MaxValidators: DefaultMaxValidators + 1}
		require.Error(t, params.Validate())
	}
}
