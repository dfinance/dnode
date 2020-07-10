// +build unit

package keeper

import (
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	tmtime "github.com/tendermint/tendermint/types/time"
	"testing"

	"github.com/dfinance/dnode/x/oracle/internal/types"
)

// TestKeeper_SetGetAsset tests adding assets to the oracle, getting assets from the store
func TestKeeper_SetGetAsset(t *testing.T) {
	t.Parallel()

	helper := getMockApp(t, 0, types.GenesisState{}, nil)
	header := abci.Header{
		Height: helper.mApp.LastBlockHeight() + 1,
		Time:   tmtime.Now()}
	helper.mApp.BeginBlock(abci.RequestBeginBlock{Header: header})
	ctx := helper.mApp.BaseApp.NewContext(false, header)

	ap := types.Params{
		Assets: []types.Asset{
			types.Asset{AssetCode: "tstusd", Oracles: types.Oracles{}, Active: true},
		},
	}
	helper.keeper.SetParams(ctx, ap)
	assets := helper.keeper.GetAssetParams(ctx)
	require.Equal(t, len(assets), 1)
	require.Equal(t, assets[0].AssetCode, "tstusd")

	_, found := helper.keeper.GetAsset(ctx, "tstusd")
	require.Equal(t, found, true)

	ap = types.Params{
		Assets: []types.Asset{
			types.Asset{AssetCode: "tstusd", Oracles: types.Oracles{}, Active: true},
			types.Asset{AssetCode: "tst2usd", Oracles: types.Oracles{}, Active: true},
		},
	}
	helper.keeper.SetParams(ctx, ap)
	assets = helper.keeper.GetAssetParams(ctx)
	require.Equal(t, len(assets), 2)
	require.Equal(t, assets[0].AssetCode, "tstusd")
	require.Equal(t, assets[1].AssetCode, "tst2usd")

	_, found = helper.keeper.GetAsset(ctx, "nan")
	require.Equal(t, found, false)
}

// nolint:errcheck
// TestKeeper_SetGetAsset tests adding assets to the oracle, getting assets from the store
func TestKeeper_SetAddAsset(t *testing.T) {
	t.Parallel()

	helper := getMockApp(t, 2, types.GenesisState{}, nil)
	header := abci.Header{
		Height: helper.mApp.LastBlockHeight() + 1,
		Time:   tmtime.Now()}
	helper.mApp.BeginBlock(abci.RequestBeginBlock{Header: header})
	ctx := helper.mApp.BaseApp.NewContext(false, header)

	ap := types.Params{
		Assets: []types.Asset{
			types.Asset{AssetCode: "tstusd", Oracles: types.Oracles{}, Active: true},
		},
		Nominees: []string{helper.addrs[0].String()},
	}
	helper.keeper.SetParams(ctx, ap)
	assets := helper.keeper.GetAssetParams(ctx)
	require.Equal(t, len(assets), 1)
	require.Equal(t, assets[0].AssetCode, "tstusd")

	_, found := helper.keeper.GetAsset(ctx, "tstusd")
	require.Equal(t, found, true)
	err := helper.keeper.AddAsset(ctx, helper.addrs[0].String(), "tst2usd", types.Asset{AssetCode: "tst2usd", Oracles: types.Oracles{}, Active: true})
	require.Nil(t, err)
	assets = helper.keeper.GetAssetParams(ctx)
	require.Equal(t, len(assets), 2)
	require.Equal(t, assets[0].AssetCode, "tstusd")
	require.Equal(t, assets[1].AssetCode, "tst2usd")

	helper.keeper.AddAsset(ctx, helper.addrs[1].String(), "tst3usd", types.Asset{AssetCode: "tst3usd", Oracles: types.Oracles{}, Active: true})
	assets = helper.keeper.GetAssetParams(ctx)
	require.Equal(t, len(assets), 2)
	require.Equal(t, assets[0].AssetCode, "tstusd")
	require.Equal(t, assets[1].AssetCode, "tst2usd")

	err = helper.keeper.AddOracle(ctx, helper.addrs[0].String(), "tst2usd", helper.addrs[1].Bytes())
	require.Nil(t, err)
	oracles, err := helper.keeper.GetOracles(ctx, "tst2usd")
	require.Nil(t, err)
	oracle := types.NewOracle(helper.addrs[1].Bytes())
	require.Equal(t, oracle, oracles[0])

	_, found = helper.keeper.GetAsset(ctx, "nan")
	require.Equal(t, found, false)
}
