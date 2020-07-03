// +build unit

package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Test params set / get.
func TestMSKeeper_Params(t *testing.T) {
	t.Parallel()

	input := NewTestInput(t)
	keeper, ctx := input.target, input.ctx

	keeper.SetIntervalToExecute(ctx, 123456)
	require.EqualValues(t, 123456, keeper.GetIntervalToExecute(ctx))
}
