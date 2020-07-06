// +build unit

package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPOAKeeper_ConfirmationsCounter(t *testing.T) {
	t.Parallel()

	input := NewTestInput(t)
	keeper, ctx := input.target, input.ctx

	// check initial state
	{
		amount := keeper.GetValidatorAmount(ctx)
		confirmations := keeper.GetEnoughConfirmations(ctx)
		require.EqualValues(t, 0, amount)
		require.EqualValues(t, uint16(1), confirmations)
	}

	// add 1st
	{
		require.NoError(t, keeper.AddValidator(ctx, sdkAddress1, ethAddress1))

		// check
		amount := keeper.GetValidatorAmount(ctx)
		confirmations := keeper.GetEnoughConfirmations(ctx)
		require.EqualValues(t, 1, amount)
		require.EqualValues(t, uint16(1), confirmations)
	}

	// add 2nd
	{
		require.NoError(t, keeper.AddValidator(ctx, sdkAddress2, ethAddress2))

		// check
		amount := keeper.GetValidatorAmount(ctx)
		confirmations := keeper.GetEnoughConfirmations(ctx)
		require.EqualValues(t, 2, amount)
		require.EqualValues(t, uint16(2), confirmations)
	}

	// add 3rd
	{
		require.NoError(t, keeper.AddValidator(ctx, sdkAddress3, ethAddress3))

		// check
		amount := keeper.GetValidatorAmount(ctx)
		confirmations := keeper.GetEnoughConfirmations(ctx)
		require.EqualValues(t, 3, amount)
		require.EqualValues(t, uint16(2), confirmations)
	}

	// add 4th
	{
		require.NoError(t, keeper.AddValidator(ctx, sdkAddress4, ethAddress4))

		// check
		amount := keeper.GetValidatorAmount(ctx)
		confirmations := keeper.GetEnoughConfirmations(ctx)
		require.EqualValues(t, 4, amount)
		require.EqualValues(t, uint16(3), confirmations)
	}

	// remove one
	{
		require.NoError(t, keeper.RemoveValidator(ctx, sdkAddress1))

		// check
		amount := keeper.GetValidatorAmount(ctx)
		confirmations := keeper.GetEnoughConfirmations(ctx)
		require.EqualValues(t, 3, amount)
		require.EqualValues(t, uint16(2), confirmations)
	}
}
