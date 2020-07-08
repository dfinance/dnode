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
		params := Params{IntervalToExecute: MinIntervalToExecute}
		require.NoError(t, params.Validate())
	}

	// fail
	{
		params := Params{IntervalToExecute: MinIntervalToExecute - 1}
		require.Error(t, params.Validate())
	}
}
