// +build unit

package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dfinance/dnode/x/oracle/internal/types"
)

// Check AddOracle method with various sets of arguments.
func TestOracleKeeper_AddOracle(t *testing.T) {
	t.Parallel()

	input := NewTestInput(t)
	keeper := input.keeper
	wrongNominee := input.addresses[0].String()
	address := input.addresses[1].Bytes()

	// Add oracles ok
	{
		err := keeper.AddOracle(input.ctx, input.stdNominee, input.stdAssetCode, address)
		require.Nil(t, err)
	}

	// double add
	{
		err := keeper.AddOracle(input.ctx, input.stdNominee, input.stdAssetCode, address)
		require.Error(t, err)
	}

	// wrong nominee
	{
		err := keeper.AddOracle(input.ctx, wrongNominee, input.stdAssetCode, address)
		require.Error(t, err)
	}

	// asset code does not exist
	{
		err := keeper.AddOracle(input.ctx, input.stdNominee, "eth_btc", address)
		require.Error(t, err)
	}
}

// Check SetOracle method with various sets of arguments.
func TestOracleKeeper_SetOracle(t *testing.T) {
	t.Parallel()

	input := NewTestInput(t)
	keeper := input.keeper
	oracleMock := types.NewOracle(input.addresses[1])

	// Set oracles
	{
		err := keeper.SetOracles(input.ctx, input.stdNominee, input.stdAssetCode, []types.Oracle{oracleMock})
		require.Nil(t, err)
	}

	// Set oracles, wrong nominee
	{
		err := keeper.SetOracles(input.ctx, input.addresses[0].String(), input.stdAssetCode, []types.Oracle{oracleMock})
		require.Error(t, err)
	}

	// Set oracles, asset code does not exist
	{
		err := keeper.SetOracles(input.ctx, input.stdNominee, "btc_eth", []types.Oracle{oracleMock})
		require.Error(t, err)
	}
}
