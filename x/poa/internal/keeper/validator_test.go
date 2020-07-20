// +build unit

package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/dfinance/dnode/x/poa/internal/types"
)

// Test adding a new validator, validator getters and check all resources created.
func TestPOAKeeper_AddValidator(t *testing.T) {
	t.Parallel()

	input := NewTestInput(t)
	keeper, ctx := input.target, input.ctx

	// ok
	{
		// check validator doesn't exist
		{
			require.False(t, keeper.HasValidator(ctx, sdkAddress1))

			_, err := keeper.GetValidator(ctx, sdkAddress1)
			require.Error(t, err)
		}

		// check resources initial states
		{
			require.EqualValues(t, 0, keeper.GetValidatorAmount(ctx))
			require.Empty(t, keeper.GetValidators(ctx))
		}

		// add
		{
			err := keeper.AddValidator(ctx, sdkAddress1, ethAddress1)
			require.NoError(t, err)
		}

		// check exists
		{
			require.True(t, keeper.HasValidator(ctx, sdkAddress1))

			validator, err := keeper.GetValidator(ctx, sdkAddress1)
			require.NoError(t, err)

			require.Equal(t, sdkAddress1.String(), validator.Address.String())
			require.Equal(t, ethAddress1, validator.EthAddress)
		}

		// check resources updated
		{
			require.EqualValues(t, 1, keeper.GetValidatorAmount(ctx))

			validators := keeper.GetValidators(ctx)
			require.Len(t, validators, 1)
			require.Equal(t, sdkAddress1.String(), validators[0].Address.String())
			require.Equal(t, ethAddress1, validators[0].EthAddress)
		}
	}

	// fail: invalid validator
	{
		err := keeper.AddValidator(ctx, sdk.AccAddress(""), ethAddress1)
		require.Error(t, err)
	}

	// fail: already exists
	{
		err := keeper.AddValidator(ctx, sdkAddress1, ethAddress1)
		require.Error(t, err)
	}

	// fail: max validators reached
	{
		keeper.setParams(ctx, types.Params{
			MaxValidators: 1,
			MinValidators: 1,
		})

		err := keeper.AddValidator(ctx, sdkAddress2, ethAddress2)
		require.Error(t, err)
	}
}

// Test removing validator and check all resources updated.
func TestPOAKeeper_RemoveValidator(t *testing.T) {
	t.Parallel()

	input := NewTestInput(t)
	keeper, ctx := input.target, input.ctx

	// set mock params
	{
		keeper.setParams(ctx, types.Params{
			MaxValidators: 3,
			MinValidators: 1,
		})
	}

	// ok
	{
		// add the 1st one
		{
			err := keeper.AddValidator(ctx, sdkAddress1, ethAddress1)
			require.NoError(t, err)
		}

		// add the 2nd one
		{
			err := keeper.AddValidator(ctx, sdkAddress2, ethAddress2)
			require.NoError(t, err)
		}

		// remove
		{
			err := keeper.RemoveValidator(ctx, sdkAddress2)
			require.NoError(t, err)

			require.False(t, keeper.HasValidator(ctx, sdkAddress2))

			_, getErr := keeper.GetValidator(ctx, sdkAddress2)
			require.Error(t, getErr)
		}

		// check resources updated
		{
			require.EqualValues(t, 1, keeper.GetValidatorAmount(ctx))

			validators := keeper.GetValidators(ctx)
			require.Len(t, validators, 1)
			require.Equal(t, sdkAddress1.String(), validators[0].Address.String())
			require.Equal(t, ethAddress1, validators[0].EthAddress)
		}
	}

	// fail: non-existing
	{
		err := keeper.RemoveValidator(ctx, sdkAddress2)
		require.Error(t, err)
	}

	// fail: min validators reached
	{
		err := keeper.RemoveValidator(ctx, sdkAddress1)
		require.Error(t, err)
	}

	// ok: removing the last one
	{
		keeper.setParams(ctx, types.Params{
			MaxValidators: 3,
			MinValidators: 0,
		})

		require.NoError(t, keeper.RemoveValidator(ctx, sdkAddress1))

		require.False(t, keeper.HasValidator(ctx, sdkAddress1))

		_, getErr := keeper.GetValidator(ctx, sdkAddress1)
		require.Error(t, getErr)

		// check resources updated
		{
			require.EqualValues(t, 0, keeper.GetValidatorAmount(ctx))

			validators := keeper.GetValidators(ctx)
			require.Len(t, validators, 0)
		}
	}
}

// Test replacing validator and check all resources updated.
func TestPOAKeeper_ReplaceValidator(t *testing.T) {
	t.Parallel()

	input := NewTestInput(t)
	keeper, ctx := input.target, input.ctx

	// ok
	{
		// add the 1st one
		{
			err := keeper.AddValidator(ctx, sdkAddress1, ethAddress1)
			require.NoError(t, err)
		}

		// add the 2nd one
		{
			err := keeper.AddValidator(ctx, sdkAddress2, ethAddress2)
			require.NoError(t, err)
		}

		// replace
		{
			err := keeper.ReplaceValidator(ctx, sdkAddress2, sdkAddress3, ethAddress3)
			require.NoError(t, err)
		}

		// check replaced
		{
			require.False(t, keeper.HasValidator(ctx, sdkAddress2))
			require.True(t, keeper.HasValidator(ctx, sdkAddress3))

			_, oldErr := keeper.GetValidator(ctx, sdkAddress2)
			require.Error(t, oldErr)

			validator, newErr := keeper.GetValidator(ctx, sdkAddress3)
			require.NoError(t, newErr)

			require.Equal(t, sdkAddress3.String(), validator.Address.String())
			require.Equal(t, ethAddress3, validator.EthAddress)
		}

		// check resources
		{
			require.EqualValues(t, 2, keeper.GetValidatorAmount(ctx))

			validators := keeper.GetValidators(ctx)
			require.Len(t, validators, 2)
			require.Equal(t, sdkAddress1.String(), validators[0].Address.String())
			require.Equal(t, ethAddress1, validators[0].EthAddress)
			require.Equal(t, sdkAddress3.String(), validators[1].Address.String())
			require.Equal(t, ethAddress3, validators[1].EthAddress)
		}
	}

	// fail: non-existing old
	{
		err := keeper.ReplaceValidator(ctx, sdkAddress2, sdkAddress4, ethAddress4)
		require.Error(t, err)
	}
}
