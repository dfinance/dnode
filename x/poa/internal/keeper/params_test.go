// +build unit

package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dfinance/dnode/x/poa/internal/types"
)

// Check params setters / getters.
func TestPOAKeeper_Params(t *testing.T) {
	t.Parallel()

	input := NewTestInput(t)
	keeper, ctx := input.target, input.ctx

	inParams := types.Params{
		MaxValidators: 10,
		MinValidators: 5,
	}

	keeper.setParams(ctx, inParams)

	outParams := keeper.GetParams(ctx)
	require.EqualValues(t, 10, outParams.MaxValidators)
	require.EqualValues(t, 5, outParams.MinValidators)

	require.EqualValues(t, 10, keeper.GetMaxValidators(ctx))
	require.EqualValues(t, 5, keeper.GetMinValidators(ctx))
}
