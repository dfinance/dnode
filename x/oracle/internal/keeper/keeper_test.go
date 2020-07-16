// +build unit

package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Check IsNominee validator method.
func TestOracleKeeper_IsNominee(t *testing.T) {
	t.Parallel()

	input := NewTestInput(t)
	keeper := input.keeper
	ctx := input.ctx

	// is nominee valid
	{
		err := keeper.IsNominee(ctx, input.stdNominee)
		require.NoError(t, err)
	}

	// is nominee false
	{
		err := keeper.IsNominee(ctx, "someNominee")
		require.Error(t, err)
	}
}
