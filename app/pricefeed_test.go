// +build unit

package app

import (
	"fmt"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	"github.com/dfinance/dnode/x/pricefeed"
)

const (
	queryOracleGetCurrentPricePathFmt = "/custom/pricefeed/price/%s"
	queryOracleGetRawPricesPathFmt    = "/custom/pricefeed/rawprices/%s/%d"
	queryOracleGetAssetsPath          = "/custom/pricefeed/assets"
)

func Test_PriceFeedQueries(t *testing.T) {
	t.Parallel()
	app, server := newTestDnApp()
	defer app.CloseConnections()
	defer server.Stop()

	genAccs, genAddrs, _, genPrivKeys := CreateGenAccounts(7, GenDefCoins(t))
	_, err := setGenesis(t, app, genAccs)
	require.NoError(t, err)

	assetCodePrefix := "asset_"

	// set params (add assets with price feeds / nominees)
	{
		app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{ChainID: chainID, Height: app.LastBlockHeight() + 1}})

		ctx := GetContext(app, false)
		ap := pricefeed.Params{
			Assets: pricefeed.Assets{
				pricefeed.Asset{
					AssetCode:  assetCodePrefix + "0",
					PriceFeeds: pricefeed.PriceFeeds{{Address: genAddrs[0]}, {Address: genAddrs[1]}, {Address: genAddrs[2]}},
					Active:     true,
				},
				pricefeed.Asset{
					AssetCode:  assetCodePrefix + "1",
					PriceFeeds: pricefeed.PriceFeeds{{Address: genAddrs[1]}},
					Active:     true,
				},
			},
			Nominees: []string{genAddrs[0].String(), genAddrs[1].String()},
		}
		app.pricefeedKeeper.SetParams(ctx, ap)

		app.EndBlock(abci.RequestEndBlock{})
		app.Commit()
	}

	// getAssets query check
	{
		response := pricefeed.Assets{}
		CheckRunQuery(t, app, nil, queryOracleGetAssetsPath, &response)
		require.Len(t, response, 2)

		require.Equal(t, assetCodePrefix+"0", response[0].AssetCode)
		require.True(t, response[0].Active)
		require.Len(t, response[0].PriceFeeds, 3)
		require.Equal(t, response[0].PriceFeeds[0].Address, genAddrs[0])
		require.Equal(t, response[0].PriceFeeds[1].Address, genAddrs[1])
		require.Equal(t, response[0].PriceFeeds[2].Address, genAddrs[2])

		require.Equal(t, assetCodePrefix+"1", response[1].AssetCode)
		require.True(t, response[1].Active)
		require.Len(t, response[1].PriceFeeds, 1)
		require.Equal(t, response[1].PriceFeeds[0].Address, genAddrs[1])
	}

	assetCode := assetCodePrefix + "0"

	// getCurrentPrice query check (no inputs yet)
	{
		response := pricefeed.CurrentPrice{}
		CheckRunQuery(t, app, nil, fmt.Sprintf(queryOracleGetCurrentPricePathFmt, assetCode), &response)
		require.Empty(t, response.AssetCode)
		require.True(t, response.Price.IsZero())
		//require.True(t, response.ReceivedAt.IsZero())
	}

	now := time.Now()
	priceValues := []sdk.Int{sdk.NewInt(1000), sdk.NewInt(2000), sdk.NewInt(1500)}
	priceTimestamps := []time.Time{now.Add(1 * time.Second), now.Add(2 * time.Second), now.Add(3 * time.Second)}

	// post prices
	{
		app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{ChainID: chainID, Height: app.LastBlockHeight() + 1}})

		{
			senderAcc, senderPrivKey := GetAccount(app, genAddrs[0]), genPrivKeys[0]

			msg := pricefeed.MsgPostPrice{
				From:       senderAcc.GetAddress(),
				AssetCode:  assetCode,
				Price:      priceValues[0],
				ReceivedAt: priceTimestamps[0],
			}

			tx := genTx([]sdk.Msg{msg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
			_, res, err := app.Deliver(tx)
			require.NoError(t, err, ResultErrorMsg(res, err))
		}
		{
			senderAcc, senderPrivKey := GetAccount(app, genAddrs[1]), genPrivKeys[1]

			msg := pricefeed.MsgPostPrice{
				From:       senderAcc.GetAddress(),
				AssetCode:  assetCode,
				Price:      priceValues[1],
				ReceivedAt: priceTimestamps[1],
			}

			tx := genTx([]sdk.Msg{msg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
			_, res, err := app.Deliver(tx)
			require.NoError(t, err, ResultErrorMsg(res, err))
		}
		{
			senderAcc, senderPrivKey := GetAccount(app, genAddrs[2]), genPrivKeys[2]

			msg := pricefeed.MsgPostPrice{
				From:       senderAcc.GetAddress(),
				AssetCode:  assetCode,
				Price:      priceValues[2],
				ReceivedAt: priceTimestamps[2],
			}

			tx := genTx([]sdk.Msg{msg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
			_, res, err := app.Deliver(tx)
			require.NoError(t, err, ResultErrorMsg(res, err))
		}

		// getRawPrices query check (before BlockEnd they shouldn't exist)
		{
			response := pricefeed.QueryRawPricesResp{}
			CheckRunQuery(t, app, nil, fmt.Sprintf(queryOracleGetRawPricesPathFmt, assetCode, GetContext(app, true).BlockHeight()), &response)
			require.Len(t, response, 0)
		}

		app.EndBlock(abci.RequestEndBlock{})
		app.Commit()
	}

	// getRawPrices query check (after BlockEnd they should exist for the previous block)
	{
		response := pricefeed.QueryRawPricesResp{}
		CheckRunQuery(t, app, nil, fmt.Sprintf(queryOracleGetRawPricesPathFmt, assetCode, GetContext(app, true).BlockHeight()-1), &response)
		require.Len(t, response, 3)
		for i, rawPrice := range response {
			require.True(t, priceValues[i].Equal(rawPrice.Price))
			require.True(t, priceTimestamps[i].Equal(rawPrice.ReceivedAt))
			require.Equal(t, assetCode, rawPrice.AssetCode)
			require.Equal(t, genAddrs[i], rawPrice.PriceFeedAddress)
		}
	}

	// getCurrentPrice query check (value should be calculated after BlockEnd)
	{
		response := pricefeed.CurrentPrice{}
		CheckRunQuery(t, app, nil, fmt.Sprintf(queryOracleGetCurrentPricePathFmt, assetCode), &response)
		require.Equal(t, assetCode, response.AssetCode)
		require.True(t, response.Price.Equal(priceValues[2]))
		require.True(t, response.ReceivedAt.Equal(priceTimestamps[2]))
	}
}

func Test_OracleAddOracle(t *testing.T) {
	t.Parallel()
	app, server := newTestDnApp()
	defer app.CloseConnections()
	defer server.Stop()

	genAccs, genAddrs, _, genPrivKeys := CreateGenAccounts(7, GenDefCoins(t))
	_, err := setGenesis(t, app, genAccs)
	require.NoError(t, err)

	nomineeAddr, nomineePrivKey, assetCode := genAddrs[0], genPrivKeys[0], "dn2dn"

	newPriceFeedAcc1, err := sdk.AccAddressFromHex(secp256k1.GenPrivKey().PubKey().Address().String())
	require.NoError(t, err)
	newPriceFeedAcc2, err := sdk.AccAddressFromHex(secp256k1.GenPrivKey().PubKey().Address().String())
	require.NoError(t, err)

	// set params (add asset / nominees)
	{
		app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{ChainID: chainID, Height: app.LastBlockHeight() + 1}})

		ctx := GetContext(app, false)
		ap := pricefeed.Params{
			Assets: pricefeed.Assets{
				pricefeed.Asset{AssetCode: assetCode, PriceFeeds: pricefeed.PriceFeeds{}, Active: true},
			},
			Nominees: []string{nomineeAddr.String()},
		}
		app.pricefeedKeeper.SetParams(ctx, ap)
		assets := app.pricefeedKeeper.GetAssetParams(ctx)
		require.Equal(t, len(assets), 1)
		require.Equal(t, assets[0].AssetCode, assetCode)

		app.EndBlock(abci.RequestEndBlock{})
		app.Commit()
	}

	// add price feed to the asset (1st)
	{
		msg := pricefeed.MsgAddPriceFeed{
			PriceFeed: newPriceFeedAcc1,
			Nominee:   nomineeAddr,
			Denom:     assetCode,
		}
		senderAcc := GetAccountCheckTx(app, nomineeAddr)

		tx := genTx([]sdk.Msg{msg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, nomineePrivKey)
		CheckDeliverTx(t, app, tx)
	}

	// add price feed to the asset (2nd)
	{
		msg := pricefeed.MsgAddPriceFeed{
			PriceFeed: newPriceFeedAcc2,
			Nominee:   nomineeAddr,
			Denom:     assetCode,
		}
		senderAcc := GetAccountCheckTx(app, nomineeAddr)

		tx := genTx([]sdk.Msg{msg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, nomineePrivKey)
		CheckDeliverTx(t, app, tx)
	}

	// check price feeds added to the asset
	{
		pricefeeds, err := app.pricefeedKeeper.GetPriceFeeds(GetContext(app, true), assetCode)
		require.NoError(t, err)
		require.Len(t, pricefeeds, 2)
	}

	// check the 1st price feed
	{
		pricefeedObj, err := app.pricefeedKeeper.GetPriceFeed(GetContext(app, true), assetCode, newPriceFeedAcc1)
		require.NoError(t, err)
		require.True(t, pricefeedObj.Address.Equals(newPriceFeedAcc1))
	}

	// check the 2nd price feed
	{
		pricefeedObj, err := app.pricefeedKeeper.GetPriceFeed(GetContext(app, true), assetCode, newPriceFeedAcc2)
		require.NoError(t, err)
		require.True(t, pricefeedObj.Address.Equals(newPriceFeedAcc2))
	}
}

func Test_OracleSetOracles(t *testing.T) {
	t.Parallel()
	app, server := newTestDnApp()
	defer app.CloseConnections()
	defer server.Stop()

	genAccs, genAddrs, _, genPrivKeys := CreateGenAccounts(7, GenDefCoins(t))
	_, err := setGenesis(t, app, genAccs)
	require.NoError(t, err)

	nomineeAddr, nomineePrivKey, assetCode := genAddrs[0], genPrivKeys[0], "db2db"

	newPriceFeedAcc1, err := sdk.AccAddressFromHex(secp256k1.GenPrivKey().PubKey().Address().String())
	require.NoError(t, err)
	newPriceFeedAcc2, err := sdk.AccAddressFromHex(secp256k1.GenPrivKey().PubKey().Address().String())
	require.NoError(t, err)
	newOracleAcc3, err := sdk.AccAddressFromHex(secp256k1.GenPrivKey().PubKey().Address().String())
	require.NoError(t, err)
	setPriceFeedAccs := []sdk.AccAddress{newPriceFeedAcc2, newOracleAcc3}

	// set params (add asset / nominees)
	{
		app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{ChainID: chainID, Height: app.LastBlockHeight() + 1}})

		ctx := GetContext(app, false)
		ap := pricefeed.Params{
			Assets: pricefeed.Assets{
				pricefeed.Asset{AssetCode: assetCode, PriceFeeds: pricefeed.PriceFeeds{}, Active: true},
			},
			Nominees: []string{nomineeAddr.String()},
		}
		app.pricefeedKeeper.SetParams(ctx, ap)
		assets := app.pricefeedKeeper.GetAssetParams(ctx)
		require.Equal(t, len(assets), 1)
		require.Equal(t, assets[0].AssetCode, assetCode)

		app.EndBlock(abci.RequestEndBlock{})
		app.Commit()
	}

	// add price feed to the asset
	{
		msg := pricefeed.MsgAddPriceFeed{
			PriceFeed: newPriceFeedAcc1,
			Nominee:   nomineeAddr,
			Denom:     assetCode,
		}
		senderAcc := GetAccountCheckTx(app, nomineeAddr)

		tx := genTx([]sdk.Msg{msg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, nomineePrivKey)
		CheckDeliverTx(t, app, tx)
	}

	// check setting price feeds to the asset (rewrite the one added before)
	{
		msg := pricefeed.MsgSetPriceFeeds{
			PriceFeeds: pricefeed.PriceFeeds{{Address: setPriceFeedAccs[0]}, {Address: setPriceFeedAccs[1]}},
			Nominee:    nomineeAddr,
			Denom:      assetCode,
		}
		senderAcc := GetAccountCheckTx(app, nomineeAddr)

		tx := genTx([]sdk.Msg{msg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, nomineePrivKey)
		CheckDeliverTx(t, app, tx)

		pricefeeds, err := app.pricefeedKeeper.GetPriceFeeds(GetContext(app, true), assetCode)
		require.NoError(t, err)

		require.Len(t, pricefeeds, len(setPriceFeedAccs))
		for i, acc := range setPriceFeedAccs {
			pricefeedObj := pricefeeds[i]
			require.True(t, acc.Equals(pricefeedObj.Address))
		}
	}
}

func Test_OracleAddAsset(t *testing.T) {
	t.Parallel()
	app, server := newTestDnApp()
	defer app.CloseConnections()
	defer server.Stop()

	genAccs, genAddrs, _, genPrivKeys := CreateGenAccounts(7, GenDefCoins(t))
	_, err := setGenesis(t, app, genAccs)
	require.NoError(t, err)

	nomineeAddr, nomineePrivKey, assetCode := genAddrs[0], genPrivKeys[0], "dn2dn"

	newPriceFeedAcc1, err := sdk.AccAddressFromHex(secp256k1.GenPrivKey().PubKey().Address().String())
	require.NoError(t, err)
	newPriceFeedAcc2, err := sdk.AccAddressFromHex(secp256k1.GenPrivKey().PubKey().Address().String())
	require.NoError(t, err)

	// set params (add asset with price feed / nominees)
	{
		app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{ChainID: chainID, Height: app.LastBlockHeight() + 1}})

		ctx := GetContext(app, false)
		ap := pricefeed.Params{
			Assets: pricefeed.Assets{
				pricefeed.Asset{AssetCode: assetCode, PriceFeeds: pricefeed.PriceFeeds{{Address: newPriceFeedAcc1}}, Active: true},
			},
			Nominees: []string{nomineeAddr.String()},
		}
		app.pricefeedKeeper.SetParams(ctx, ap)
		assets := app.pricefeedKeeper.GetAssetParams(ctx)
		require.Equal(t, len(assets), 1)
		require.Equal(t, assets[0].AssetCode, assetCode)
		require.Len(t, assets[0].PriceFeeds, 1)

		app.EndBlock(abci.RequestEndBlock{})
		app.Commit()
	}

	// check adding new asset with existing nominees
	newAssetCode := "dn2test"
	{
		msg := pricefeed.MsgAddAsset{
			Nominee: nomineeAddr,
			Denom:   newAssetCode,
			Asset:   pricefeed.NewAsset(newAssetCode, pricefeed.PriceFeeds{{Address: newPriceFeedAcc1}, {Address: newPriceFeedAcc2}}, true),
		}
		senderAcc := GetAccountCheckTx(app, genAccs[0].Address)

		tx := genTx([]sdk.Msg{msg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, nomineePrivKey)
		CheckDeliverTx(t, app, tx)

		// check the new one added
		newAsset, found := app.pricefeedKeeper.GetAsset(GetContext(app, true), newAssetCode)
		require.True(t, found)
		require.Equal(t, newAssetCode, newAsset.AssetCode)
		require.Equal(t, true, newAsset.Active)
		require.Len(t, newAsset.PriceFeeds, 2)
		require.True(t, newAsset.PriceFeeds[0].Address.Equals(newPriceFeedAcc1))
		require.True(t, newAsset.PriceFeeds[1].Address.Equals(newPriceFeedAcc2))

		// check the old one still exists
		oldAsset, found := app.pricefeedKeeper.GetAsset(GetContext(app, true), assetCode)
		require.True(t, found)
		require.Equal(t, assetCode, oldAsset.AssetCode)
		require.Equal(t, true, oldAsset.Active)
		require.Len(t, oldAsset.PriceFeeds, 1)
		require.True(t, oldAsset.PriceFeeds[0].Address.Equals(newPriceFeedAcc1))
	}

	// check adding new asset with non-existing nominees
	{
		msg := pricefeed.MsgAddPriceFeed{
			PriceFeed: newPriceFeedAcc1,
			Nominee:   newPriceFeedAcc2,
			Denom:     "non-existing-asset",
		}
		senderAcc := GetAccountCheckTx(app, nomineeAddr)

		tx := genTx([]sdk.Msg{msg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, nomineePrivKey)
		// TODO: add specific error check
		CheckDeliverErrorTx(t, app, tx)
	}
}

func Test_OracleSetAsset(t *testing.T) {
	t.Parallel()
	app, server := newTestDnApp()
	defer app.CloseConnections()
	defer server.Stop()

	genAccs, genAddrs, _, genPrivKeys := CreateGenAccounts(7, GenDefCoins(t))
	_, err := setGenesis(t, app, genAccs)
	require.NoError(t, err)

	nomineeAddr, nomineePrivKey, assetCode := genAddrs[0], genPrivKeys[0], "dn2dn"

	newOracleAcc1, err := sdk.AccAddressFromHex(secp256k1.GenPrivKey().PubKey().Address().String())
	require.NoError(t, err)
	newOracleAcc2, err := sdk.AccAddressFromHex(secp256k1.GenPrivKey().PubKey().Address().String())
	require.NoError(t, err)

	// set params (add asset with price feed / nominees)
	{
		app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{ChainID: chainID, Height: app.LastBlockHeight() + 1}})

		ctx := GetContext(app, false)
		ap := pricefeed.Params{
			Assets: pricefeed.Assets{
				pricefeed.Asset{AssetCode: assetCode, PriceFeeds: pricefeed.PriceFeeds{{Address: newOracleAcc1}}, Active: true},
			},
			Nominees: []string{nomineeAddr.String()},
		}
		app.pricefeedKeeper.SetParams(ctx, ap)
		assets := app.pricefeedKeeper.GetAssetParams(ctx)
		require.Equal(t, len(assets), 1)
		require.Equal(t, assets[0].AssetCode, assetCode)
		require.Len(t, assets[0].PriceFeeds, 1)

		app.EndBlock(abci.RequestEndBlock{})
		app.Commit()
	}

	// check asset created with SetParams exists
	{
		asset, found := app.pricefeedKeeper.GetAsset(GetContext(app, true), assetCode)
		require.True(t, found)
		require.Equal(t, assetCode, asset.AssetCode)
		require.True(t, asset.Active)
		require.Len(t, asset.PriceFeeds, 1)
		require.Equal(t, newOracleAcc1, asset.PriceFeeds[0].Address)
	}

	// check setting asset (updating)
	{
		updAssetCode := "dn2test1"

		msg := pricefeed.MsgSetAsset{
			Nominee: nomineeAddr,
			Denom:   assetCode,
			Asset:   pricefeed.NewAsset(updAssetCode, pricefeed.PriceFeeds{{Address: newOracleAcc1}, {Address: newOracleAcc2}}, true),
		}
		senderAcc := GetAccountCheckTx(app, nomineeAddr)

		tx := genTx([]sdk.Msg{msg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, nomineePrivKey)
		CheckDeliverTx(t, app, tx)

		asset, found := app.pricefeedKeeper.GetAsset(GetContext(app, true), updAssetCode)
		require.True(t, found)
		require.Equal(t, updAssetCode, asset.AssetCode)
		require.True(t, asset.Active)
		require.Len(t, asset.PriceFeeds, 2)
		require.True(t, asset.PriceFeeds[0].Address.Equals(newOracleAcc1))
		require.True(t, asset.PriceFeeds[1].Address.Equals(newOracleAcc2))
	}
}

func Test_OraclePostPrices(t *testing.T) {
	t.Parallel()
	app, server := newTestDnApp()
	defer app.CloseConnections()
	defer server.Stop()

	genAccs, genAddrs, _, genPrivKeys := CreateGenAccounts(7, GenDefCoins(t))
	_, err := setGenesis(t, app, genAccs)
	require.NoError(t, err)

	assetCode := "dn2dn"

	// set params (add asset with price feed 0 / nominees)
	{
		nomineeAddr := genAddrs[0]

		app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{ChainID: chainID, Height: app.LastBlockHeight() + 1}})

		ctx := GetContext(app, false)
		ap := pricefeed.Params{
			Assets: pricefeed.Assets{
				pricefeed.Asset{AssetCode: assetCode, PriceFeeds: pricefeed.PriceFeeds{{Address: genAddrs[0]}}, Active: true},
			},
			Nominees: []string{nomineeAddr.String()},
		}
		app.pricefeedKeeper.SetParams(ctx, ap)
		assets := app.pricefeedKeeper.GetAssetParams(ctx)
		require.Equal(t, len(assets), 1)
		require.Equal(t, assets[0].AssetCode, assetCode)

		app.EndBlock(abci.RequestEndBlock{})
		app.Commit()
	}

	// check posting price from non-existing price feed
	{
		senderAcc, senderPrivKey := GetAccountCheckTx(app, genAddrs[1]), genPrivKeys[1]

		msg := pricefeed.MsgPostPrice{
			From:       senderAcc.GetAddress(),
			AssetCode:  assetCode,
			Price:      sdk.OneInt(),
			ReceivedAt: time.Now(),
		}

		tx := genTx([]sdk.Msg{msg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
		// TODO: add specific error check
		CheckDeliverErrorTx(t, app, tx)
	}

	// check posting price for non-existing asset
	{
		senderAcc, senderPrivKey := GetAccountCheckTx(app, genAddrs[0]), genPrivKeys[0]

		msg := pricefeed.MsgPostPrice{
			From:       senderAcc.GetAddress(),
			AssetCode:  "non-existing-asset",
			Price:      sdk.OneInt(),
			ReceivedAt: time.Now(),
		}

		tx := genTx([]sdk.Msg{msg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
		// TODO: add specific error check
		CheckDeliverErrorTx(t, app, tx)
	}

	// set price feeds for the asset
	{
		senderAcc, senderPrivKey := GetAccountCheckTx(app, genAddrs[0]), genPrivKeys[0]

		msg := pricefeed.MsgSetPriceFeeds{
			PriceFeeds: pricefeed.PriceFeeds{{Address: genAddrs[0]}, {Address: genAddrs[1]}, {Address: genAddrs[2]}},
			Nominee:    senderAcc.GetAddress(),
			Denom:      assetCode,
		}

		tx := genTx([]sdk.Msg{msg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
		CheckDeliverTx(t, app, tx)
	}

	// check posting price few times from the same price feed
	{
		now := time.Now()
		priceAmount1, priceAmount2 := sdk.NewInt(200000000), sdk.NewInt(100000000)
		priceTimestamp1, priceTimestamp2 := now.Add(1*time.Second), now.Add(2*time.Second)

		// post prices
		app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{ChainID: chainID, Height: app.LastBlockHeight() + 1}})
		{
			senderAcc, senderPrivKey := GetAccount(app, genAddrs[0]), genPrivKeys[0]

			msg := pricefeed.MsgPostPrice{
				From:       senderAcc.GetAddress(),
				AssetCode:  assetCode,
				Price:      priceAmount1,
				ReceivedAt: priceTimestamp1,
			}

			tx := genTx([]sdk.Msg{msg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
			_, res, err := app.Deliver(tx)
			require.NoError(t, err, ResultErrorMsg(res, err))
		}
		{
			senderAcc, senderPrivKey := GetAccount(app, genAddrs[0]), genPrivKeys[0]

			msg := pricefeed.MsgPostPrice{
				From:       senderAcc.GetAddress(),
				AssetCode:  assetCode,
				Price:      priceAmount2,
				ReceivedAt: priceTimestamp2,
			}

			tx := genTx([]sdk.Msg{msg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
			_, res, err := app.Deliver(tx)
			require.NoError(t, err, ResultErrorMsg(res, err))
		}
		app.EndBlock(abci.RequestEndBlock{})
		app.Commit()

		// check the last price is the current price
		{
			price := app.pricefeedKeeper.GetCurrentPrice(GetContext(app, true), assetCode)
			require.True(t, price.Price.Equal(priceAmount2))
			require.True(t, price.ReceivedAt.Equal(priceTimestamp2))
		}

		// check rawPrices
		{
			ctx := GetContext(app, true)
			rawPrices := app.pricefeedKeeper.GetRawPrices(ctx, assetCode, ctx.BlockHeight()-1)
			require.Len(t, rawPrices, 1)
			require.True(t, priceAmount2.Equal(rawPrices[0].Price))
			require.True(t, priceTimestamp2.Equal(rawPrices[0].ReceivedAt))
			require.Equal(t, assetCode, rawPrices[0].AssetCode)
			require.Equal(t, genAddrs[0], rawPrices[0].PriceFeedAddress)
		}
	}

	// check posting prices from different price feeds
	{
		now := time.Now()
		priceValues := []sdk.Int{sdk.NewInt(200000000), sdk.NewInt(100000000), sdk.NewInt(300000000)}
		priceTimestamps := []time.Time{now.Add(1 * time.Second), now.Add(2 * time.Second), now.Add(3 * time.Second)}

		// post prices
		app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{ChainID: chainID, Height: app.LastBlockHeight() + 1}})
		{
			senderAcc, senderPrivKey := GetAccount(app, genAddrs[0]), genPrivKeys[0]

			msg := pricefeed.MsgPostPrice{
				From:       senderAcc.GetAddress(),
				AssetCode:  assetCode,
				Price:      priceValues[0],
				ReceivedAt: priceTimestamps[0],
			}

			tx := genTx([]sdk.Msg{msg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
			_, res, err := app.Deliver(tx)
			require.NoError(t, err, ResultErrorMsg(res, err))
		}
		{
			senderAcc, senderPrivKey := GetAccount(app, genAddrs[1]), genPrivKeys[1]

			msg := pricefeed.MsgPostPrice{
				From:       senderAcc.GetAddress(),
				AssetCode:  assetCode,
				Price:      priceValues[1],
				ReceivedAt: priceTimestamps[1],
			}

			tx := genTx([]sdk.Msg{msg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
			_, res, err := app.Deliver(tx)
			require.NoError(t, err, ResultErrorMsg(res, err))
		}
		{
			senderAcc, senderPrivKey := GetAccount(app, genAddrs[2]), genPrivKeys[2]

			msg := pricefeed.MsgPostPrice{
				From:       senderAcc.GetAddress(),
				AssetCode:  assetCode,
				Price:      priceValues[2],
				ReceivedAt: priceTimestamps[2],
			}

			tx := genTx([]sdk.Msg{msg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
			_, res, err := app.Deliver(tx)
			require.NoError(t, err, ResultErrorMsg(res, err))
		}
		app.EndBlock(abci.RequestEndBlock{})
		app.Commit()

		// check the last price is the median price
		{
			price := app.pricefeedKeeper.GetCurrentPrice(GetContext(app, true), assetCode)
			require.True(t, price.Price.Equal(priceValues[0]))
			require.True(t, price.ReceivedAt.Equal(priceTimestamps[0]))
		}

		// check rawPrices
		{
			ctx := GetContext(app, true)
			rawPrices := app.pricefeedKeeper.GetRawPrices(ctx, assetCode, ctx.BlockHeight()-1)
			require.Len(t, rawPrices, 3)
			for i, rawPrice := range rawPrices {
				require.True(t, priceValues[i].Equal(rawPrice.Price))
				require.True(t, priceTimestamps[i].Equal(rawPrice.ReceivedAt))
				require.Equal(t, assetCode, rawPrice.AssetCode)
				require.Equal(t, genAddrs[i], rawPrice.PriceFeedAddress)
			}
		}

		// check rawPrices from the previous block are still exist
		{
			ctx := GetContext(app, true)
			rawPrices := app.pricefeedKeeper.GetRawPrices(ctx, assetCode, ctx.BlockHeight()-2)
			require.Len(t, rawPrices, 1)
		}
	}
}
