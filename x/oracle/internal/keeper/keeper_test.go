// +build unit

package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOracleKeeper_IsNominee(t *testing.T) {
	t.Parallel()

	input := NewTestInput(t)
	keeper := input.keeper
	ctx := input.ctx

	// is nominee valid
	{
		ok := keeper.IsNominee(ctx, input.stdNominee)
		require.True(t, ok)
	}

	// is nominee false
	{
		ok := keeper.IsNominee(ctx, "someNominee")
		require.False(t, ok)
	}
}
