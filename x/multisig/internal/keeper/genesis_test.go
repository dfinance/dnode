// +build unit

package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dfinance/dnode/x/multisig/internal/types"
)

// Check genesis currencies created and params updated.
func TestMSKeeper_InitGenesis(t *testing.T) {
	t.Parallel()

	input := NewTestInput(t)
	keeper, ctx, cdc := input.target, input.ctx, input.cdc

	state := types.GenesisState{
		Parameters: types.Params{
			IntervalToExecute: 123,
		},
	}

	keeper.InitGenesis(ctx, cdc.MustMarshalJSON(state))
	require.EqualValues(t, 123, keeper.GetIntervalToExecute(ctx))
}

// Check runtime genesis export.
func TestMSKeeper_ExportGenesis(t *testing.T) {
	t.Parallel()

	input := NewTestInput(t)
	keeper, ctx, cdc := input.target, input.ctx, input.cdc

	defaultGenesis := types.DefaultGenesisState()
	keeper.InitGenesis(ctx, cdc.MustMarshalJSON(defaultGenesis))

	var exportedState types.GenesisState
	cdc.MustUnmarshalJSON(keeper.ExportGenesis(ctx), &exportedState)

	require.Equal(t, defaultGenesis.Parameters.IntervalToExecute, exportedState.Parameters.IntervalToExecute)
}
