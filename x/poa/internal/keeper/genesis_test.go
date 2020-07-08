// +build unit

package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dfinance/dnode/x/poa/internal/types"
)

// Check genesis validators created and params updated, check runtime export.
func TestPOAKeeper_InitExportGenesis(t *testing.T) {
	t.Parallel()

	input := NewTestInput(t)
	keeper, ctx, cdc := input.target, input.ctx, input.cdc

	validatorsInput := types.Validators{
		types.NewValidator(sdkAddress1, ethAddress1),
		types.NewValidator(sdkAddress2, ethAddress2),
		types.NewValidator(sdkAddress3, ethAddress3),
	}

	// init
	{
		state := types.GenesisState{
			Parameters: types.Params{
				MaxValidators: 4,
				MinValidators: 3,
			},
			Validators: validatorsInput,
		}
		keeper.InitGenesis(ctx, cdc.MustMarshalJSON(state))

		outParams := keeper.GetParams(ctx)
		require.EqualValues(t, 4, outParams.MaxValidators)
		require.EqualValues(t, 3, outParams.MinValidators)

		outValidators := keeper.GetValidators(ctx)
		require.Len(t, outValidators, len(validatorsInput))

		for i, inValidator := range validatorsInput {
			outValidator := outValidators[i]
			require.Equal(t, inValidator.Address.String(), outValidator.Address.String())
			require.Equal(t, inValidator.EthAddress, outValidator.EthAddress)
		}
	}

	// export
	{
		var exportedState types.GenesisState
		cdc.MustUnmarshalJSON(keeper.ExportGenesis(ctx), &exportedState)

		require.EqualValues(t, 4, exportedState.Parameters.MaxValidators)
		require.EqualValues(t, 3, exportedState.Parameters.MinValidators)

		require.Len(t, exportedState.Validators, len(validatorsInput))
		for i, inValidator := range validatorsInput {
			outValidator := exportedState.Validators[i]
			require.Equal(t, inValidator.Address.String(), outValidator.Address.String())
			require.Equal(t, inValidator.EthAddress, outValidator.EthAddress)
		}
	}
}
