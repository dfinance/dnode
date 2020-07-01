// +build unit

package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Test keeper store/get PathData methods.
func TestCCSKeeper_PathData(t *testing.T) {
	t.Parallel()

	input := NewTestInput(t)
	ctx, keeper := input.ctx, input.keeper

	// ok
	{
		key, inPath := []byte("key"), []byte("path")

		keeper.storePathData(ctx, key, inPath)
		outPath, ok := keeper.getPathData(ctx, key)

		require.True(t, ok)
		require.EqualValues(t, inPath, outPath)
	}

	// fail: non-existing
	{
		_, ok := keeper.getPathData(ctx, []byte("non-existing"))
		require.False(t, ok)
	}
}
