// +build unit

package keeper

import (
	dnTypes "github.com/dfinance/dnode/helpers/types"
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
		err := keeper.SetAsset(ctx, input.stdNominee, asset)
		require.Nil(t, err)
	}

	// set asset with wrong nominee
	{
		err := keeper.SetAsset(ctx, "wrongNominee", asset)
		require.Error(t, err)
	}

	// wrong asset code, doesn't exist
	{
		assetT := &asset
		asset2 := *assetT
		asset2.AssetCode = dnTypes.AssetCode("btc_eth")
		err := keeper.SetAsset(ctx, input.stdNominee, asset2)
		require.Error(t, err)
	}
}

func TestOracleKeeper_AddAsset(t *testing.T) {
	t.Parallel()

	input := NewTestInput(t)
	keeper := input.keeper
	ctx := input.ctx

	asset := types.NewAsset("btc_usdt", []types.Oracle{}, true)

	// add asset
	{
		err := keeper.AddAsset(ctx, input.stdNominee, asset)
		require.Nil(t, err)

		_, ok := keeper.GetAsset(ctx, asset.AssetCode)
		require.True(t, ok)
	}

	// add asset with wrong nominee
	{
		err := keeper.AddAsset(ctx, "wrongNominee", asset)
		require.Error(t, err)
	}

	// double add
	{
		asset2 := types.NewAsset("btc_usdt", []types.Oracle{}, true)
		err := keeper.AddAsset(ctx, input.stdNominee, asset2)
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
		err := keeper.SetAsset(ctx, input.stdNominee, asset)
		require.Nil(t, err)

		a, ok := keeper.GetAsset(ctx, input.stdAssetCode)
		require.Equal(t, true, ok)
		require.Equal(t, a.AssetCode, input.stdAssetCode)

		_, ok = keeper.GetAsset(ctx, "btc_eth")
		require.Equal(t, false, ok)
	}
}
