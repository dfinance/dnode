// +build unit

package keeper

import (
	"github.com/stretchr/testify/require"
	"testing"
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
