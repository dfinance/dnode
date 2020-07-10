// +build unit

package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dfinance/dnode/x/oracle/internal/types"
)

func TestOracleKeeper_SetAsset(t *testing.T) {
	t.Parallel()

	input := NewTestInput(t)
	keeper := input.keeper
	ctx := input.ctx

	asset := types.NewAsset(input.stdAssetCode, []types.Oracle{}, true)

	// set asset
	{
		err := keeper.SetAsset(ctx, input.stdNominee, input.stdAssetCode, asset)
		require.Nil(t, err)
	}

	// set asset with wrong nominee
	{
		err := keeper.SetAsset(ctx, "wrongNominee", input.stdAssetCode, asset)
		require.Error(t, err)
	}

	// wrong asset code, doesn't exist
	{
		err := keeper.SetAsset(ctx, input.stdNominee, "btc_eth", asset)
		require.Error(t, err)
	}
}

func TestOracleKeeper_GetAsset(t *testing.T) {
	t.Parallel()

	input := NewTestInput(t)
	keeper := input.keeper
	ctx := input.ctx

	asset := types.NewAsset(input.stdAssetCode, []types.Oracle{}, true)

	// get asset
	{
		err := keeper.SetAsset(ctx, input.stdNominee, input.stdAssetCode, asset)
		require.Nil(t, err)

		a, ok := keeper.GetAsset(ctx, input.stdAssetCode)
		require.Equal(t, true, ok)
		require.Equal(t, a.AssetCode, input.stdAssetCode)

		_, ok = keeper.GetAsset(ctx, "btc_eth")
		require.Equal(t, false, ok)
	}
}
