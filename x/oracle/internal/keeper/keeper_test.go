package keeper_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"wings-blockchain/x/oracle/internal/types"
)

// TestKeeper_SetGetAsset tests adding assets to the oracle, getting assets from the store
func TestKeeper_SetGetAsset(t *testing.T) {
	helper := getMockApp(t, 0, types.GenesisState{}, nil)
	header := abci.Header{
		Height: helper.mApp.LastBlockHeight() + 1,
		Time:   tmtime.Now()}
	helper.mApp.BeginBlock(abci.RequestBeginBlock{Header: header})
	ctx := helper.mApp.BaseApp.NewContext(false, header)

	ap := types.Params{
		Assets: []types.Asset{
			types.Asset{AssetCode: "tstusd", BaseAsset: "tst", QuoteAsset: "usd", Oracles: types.Oracles{}, Active: true},
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
			types.Asset{AssetCode: "tstusd", BaseAsset: "tst", QuoteAsset: "usd", Oracles: types.Oracles{}, Active: true},
			types.Asset{AssetCode: "tst2usd", BaseAsset: "tst2", QuoteAsset: "usd", Oracles: types.Oracles{}, Active: true},
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
	helper := getMockApp(t, 2, types.GenesisState{}, nil)
	header := abci.Header{
		Height: helper.mApp.LastBlockHeight() + 1,
		Time:   tmtime.Now()}
	helper.mApp.BeginBlock(abci.RequestBeginBlock{Header: header})
	ctx := helper.mApp.BaseApp.NewContext(false, header)

	ap := types.Params{
		Assets: []types.Asset{
			types.Asset{AssetCode: "tstusd", BaseAsset: "tst", QuoteAsset: "usd", Oracles: types.Oracles{}, Active: true},
		},
		Nominees: []string{helper.addrs[0].String()},
	}
	helper.keeper.SetParams(ctx, ap)
	assets := helper.keeper.GetAssetParams(ctx)
	require.Equal(t, len(assets), 1)
	require.Equal(t, assets[0].AssetCode, "tstusd")

	_, found := helper.keeper.GetAsset(ctx, "tstusd")
	require.Equal(t, found, true)
	err := helper.keeper.AddAsset(ctx, helper.addrs[0].String(), "tst2usd", types.Asset{AssetCode: "tst2usd", BaseAsset: "tst", QuoteAsset: "usd", Oracles: types.Oracles{}, Active: true})
	require.Nil(t, err)
	assets = helper.keeper.GetAssetParams(ctx)
	require.Equal(t, len(assets), 2)
	require.Equal(t, assets[0].AssetCode, "tstusd")
	require.Equal(t, assets[1].AssetCode, "tst2usd")

	helper.keeper.AddAsset(ctx, helper.addrs[1].String(), "tst3usd", types.Asset{AssetCode: "tst3usd", BaseAsset: "tst", QuoteAsset: "usd", Oracles: types.Oracles{}, Active: true})
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

// TestKeeper_GetSetPrice Test Posting the price by an oracle
func TestKeeper_GetSetPrice(t *testing.T) {
	helper := getMockApp(t, 2, types.GenesisState{}, nil)
	header := abci.Header{
		Height: helper.mApp.LastBlockHeight() + 1,
		Time:   tmtime.Now()}
	helper.mApp.BeginBlock(abci.RequestBeginBlock{Header: header})
	ctx := helper.mApp.BaseApp.NewContext(false, header)
	ap := types.Params{
		Assets: []types.Asset{
			types.Asset{AssetCode: "tstusd", BaseAsset: "tst", QuoteAsset: "usd", Oracles: types.Oracles{}, Active: true},
		},
	}
	helper.keeper.SetParams(ctx, ap)
	// Set price by oracle 1
	_, err := helper.keeper.SetPrice(
		ctx, helper.addrs[0], "tstusd",
		sdk.MustNewDecFromStr("0.33"),
		header.Time.Add(1*time.Hour))
	require.NoError(t, err)
	// Get raw prices
	rawPrices := helper.keeper.GetRawPrices(ctx, "tstusd")
	require.Equal(t, len(rawPrices), 1)
	require.Equal(t, rawPrices[0].Price.Equal(sdk.MustNewDecFromStr("0.33")), true)
	// Set price by oracle 2
	_, err = helper.keeper.SetPrice(
		ctx, helper.addrs[1], "tstusd",
		sdk.MustNewDecFromStr("0.35"),
		header.Time.Add(time.Hour*1))
	require.NoError(t, err)

	rawPrices = helper.keeper.GetRawPrices(ctx, "tstusd")
	require.Equal(t, len(rawPrices), 2)
	require.Equal(t, rawPrices[1].Price.Equal(sdk.MustNewDecFromStr("0.35")), true)

	// Update Price by Oracle 1
	_, err = helper.keeper.SetPrice(
		ctx, helper.addrs[0], "tstusd",
		sdk.MustNewDecFromStr("0.37"),
		header.Time.Add(time.Hour*1))
	require.NoError(t, err)
	rawPrices = helper.keeper.GetRawPrices(ctx, "tstusd")
	require.Equal(t, rawPrices[0].Price.Equal(sdk.MustNewDecFromStr("0.37")), true)
}

// nolint:errcheck
// TestKeeper_GetSetCurrentPrice Test Setting the median price of an Asset
func TestKeeper_GetSetCurrentPrice(t *testing.T) {
	helper := getMockApp(t, 4, types.GenesisState{}, nil)
	header := abci.Header{
		Height: helper.mApp.LastBlockHeight() + 1,
		Time:   tmtime.Now()}
	helper.mApp.BeginBlock(abci.RequestBeginBlock{Header: header})
	ctx := helper.mApp.BaseApp.NewContext(false, header)
	ap := types.Params{
		Assets: []types.Asset{
			types.Asset{AssetCode: "tstusd", BaseAsset: "tst", QuoteAsset: "usd", Oracles: types.Oracles{}, Active: true},
		},
	}
	helper.keeper.SetParams(ctx, ap)
	helper.keeper.SetPrice(
		ctx, helper.addrs[0], "tstusd",
		sdk.MustNewDecFromStr("0.33"),
		header.Time.Add(time.Hour*1))
	helper.keeper.SetPrice(
		ctx, helper.addrs[1], "tstusd",
		sdk.MustNewDecFromStr("0.35"),
		header.Time.Add(time.Hour*1))
	helper.keeper.SetPrice(
		ctx, helper.addrs[2], "tstusd",
		sdk.MustNewDecFromStr("0.34"),
		header.Time.Add(time.Hour*1))
	// Set current price
	err := helper.keeper.SetCurrentPrices(ctx)
	require.NoError(t, err)
	// Get Current price
	price := helper.keeper.GetCurrentPrice(ctx, "tstusd")
	require.Equal(t, price.Price.Equal(sdk.MustNewDecFromStr("0.34")), true)

	// Even number of oracles
	helper.keeper.SetPrice(
		ctx, helper.addrs[3], "tstusd",
		sdk.MustNewDecFromStr("0.36"),
		header.Time.Add(time.Hour*1))
	err = helper.keeper.SetCurrentPrices(ctx)
	require.NoError(t, err)
	price = helper.keeper.GetCurrentPrice(ctx, "tstusd")
	require.Equal(t, price.Price.Equal(sdk.MustNewDecFromStr("0.345")), true)

}
