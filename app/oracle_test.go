package app

import (
	"fmt"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	"github.com/dfinance/dnode/x/oracle"
)

const (
	queryOracleGetCurrentPricePathFmt = "/custom/oracle/price/%s"
	queryOracleGetRawPricesPathFmt    = "/custom/oracle/rawprices/%s/%d"
	queryOracleGetAssetsPath          = "/custom/oracle/assets"
)

func Test_OracleQueries(t *testing.T) {
	app, server := newTestDnApp()
	defer app.CloseConnections()
	defer server.Stop()

	genAccs, genAddrs, _, genPrivKeys := CreateGenAccounts(7, GenDefCoins(t))
	_, err := setGenesis(t, app, genAccs)
	require.NoError(t, err)

	assetCodePrefix := "asset_"

	// set params (add assets with oracles / nominees)
	{
		app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{ChainID: chainID, Height: app.LastBlockHeight() + 1}})

		ctx := GetContext(app, false)
		ap := oracle.Params{
			Assets: oracle.Assets{
				oracle.Asset{
					AssetCode: assetCodePrefix + "0",
					Oracles:   oracle.Oracles{{Address: genAddrs[0]}, {Address: genAddrs[1]}, {Address: genAddrs[2]}},
					Active:    true,
				},
				oracle.Asset{
					AssetCode: assetCodePrefix + "1",
					Oracles:   oracle.Oracles{{Address: genAddrs[1]}},
					Active:    true,
				},
			},
			Nominees: []string{genAddrs[0].String(), genAddrs[1].String()},
		}
		app.oracleKeeper.SetParams(ctx, ap)

		app.EndBlock(abci.RequestEndBlock{})
		app.Commit()
	}

	// getAssets query check
	{
		response := oracle.Assets{}
		CheckRunQuery(t, app, nil, queryOracleGetAssetsPath, &response)
		require.Len(t, response, 2)

		require.Equal(t, assetCodePrefix+"0", response[0].AssetCode)
		require.True(t, response[0].Active)
		require.Len(t, response[0].Oracles, 3)
		require.Equal(t, response[0].Oracles[0].Address, genAddrs[0])
		require.Equal(t, response[0].Oracles[1].Address, genAddrs[1])
		require.Equal(t, response[0].Oracles[2].Address, genAddrs[2])

		require.Equal(t, assetCodePrefix+"1", response[1].AssetCode)
		require.True(t, response[1].Active)
		require.Len(t, response[1].Oracles, 1)
		require.Equal(t, response[1].Oracles[0].Address, genAddrs[1])
	}

	assetCode := assetCodePrefix + "0"

	// getCurrentPrice query check (no inputs yet)
	{
		response := oracle.CurrentPrice{}
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

			msg := oracle.MsgPostPrice{
				From:       senderAcc.GetAddress(),
				AssetCode:  assetCode,
				Price:      priceValues[0],
				ReceivedAt: priceTimestamps[0],
			}

			tx := genTx([]sdk.Msg{msg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
			res := app.Deliver(tx)
			require.True(t, res.IsOK())
		}
		{
			senderAcc, senderPrivKey := GetAccount(app, genAddrs[1]), genPrivKeys[1]

			msg := oracle.MsgPostPrice{
				From:       senderAcc.GetAddress(),
				AssetCode:  assetCode,
				Price:      priceValues[1],
				ReceivedAt: priceTimestamps[1],
			}

			tx := genTx([]sdk.Msg{msg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
			res := app.Deliver(tx)
			require.True(t, res.IsOK(), res.Log)
		}
		{
			senderAcc, senderPrivKey := GetAccount(app, genAddrs[2]), genPrivKeys[2]

			msg := oracle.MsgPostPrice{
				From:       senderAcc.GetAddress(),
				AssetCode:  assetCode,
				Price:      priceValues[2],
				ReceivedAt: priceTimestamps[2],
			}

			tx := genTx([]sdk.Msg{msg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
			res := app.Deliver(tx)
			require.True(t, res.IsOK())
		}

		// getRawPrices query check (before BlockEnd they shouldn't exist)
		{
			response := oracle.QueryRawPricesResp{}
			CheckRunQuery(t, app, nil, fmt.Sprintf(queryOracleGetRawPricesPathFmt, assetCode, GetContext(app, true).BlockHeight()), &response)
			require.Len(t, response, 0)
		}

		app.EndBlock(abci.RequestEndBlock{})
		app.Commit()
	}

	// getRawPrices query check (after BlockEnd they should exist for the previous block)
	{
		response := oracle.QueryRawPricesResp{}
		CheckRunQuery(t, app, nil, fmt.Sprintf(queryOracleGetRawPricesPathFmt, assetCode, GetContext(app, true).BlockHeight()-1), &response)
		require.Len(t, response, 3)
		for i, rawPrice := range response {
			require.True(t, priceValues[i].Equal(rawPrice.Price))
			require.True(t, priceTimestamps[i].Equal(rawPrice.ReceivedAt))
			require.Equal(t, assetCode, rawPrice.AssetCode)
			require.Equal(t, genAddrs[i], rawPrice.OracleAddress)
		}
	}

	// getCurrentPrice query check (value should be calculated after BlockEnd)
	{
		response := oracle.CurrentPrice{}
		CheckRunQuery(t, app, nil, fmt.Sprintf(queryOracleGetCurrentPricePathFmt, assetCode), &response)
		require.Equal(t, assetCode, response.AssetCode)
		require.True(t, response.Price.Equal(priceValues[2]))
		require.True(t, response.ReceivedAt.Equal(priceTimestamps[2]))
	}
}

func Test_OracleAddOracle(t *testing.T) {
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

	// set params (add asset / nominees)
	{
		app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{ChainID: chainID, Height: app.LastBlockHeight() + 1}})

		ctx := GetContext(app, false)
		ap := oracle.Params{
			Assets: oracle.Assets{
				oracle.Asset{AssetCode: assetCode, Oracles: oracle.Oracles{}, Active: true},
			},
			Nominees: []string{nomineeAddr.String()},
		}
		app.oracleKeeper.SetParams(ctx, ap)
		assets := app.oracleKeeper.GetAssetParams(ctx)
		require.Equal(t, len(assets), 1)
		require.Equal(t, assets[0].AssetCode, assetCode)

		app.EndBlock(abci.RequestEndBlock{})
		app.Commit()
	}

	// add oracle to the asset (1st)
	{
		msg := oracle.MsgAddOracle{
			Oracle:  newOracleAcc1,
			Nominee: nomineeAddr,
			Denom:   assetCode,
		}
		senderAcc := GetAccountCheckTx(app, nomineeAddr)

		tx := genTx([]sdk.Msg{msg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, nomineePrivKey)
		CheckDeliverTx(t, app, tx)
	}

	// add oracle to the asset (2nd)
	{
		msg := oracle.MsgAddOracle{
			Oracle:  newOracleAcc2,
			Nominee: nomineeAddr,
			Denom:   assetCode,
		}
		senderAcc := GetAccountCheckTx(app, nomineeAddr)

		tx := genTx([]sdk.Msg{msg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, nomineePrivKey)
		CheckDeliverTx(t, app, tx)
	}

	// check oracles added to the asset
	{
		oracles, err := app.oracleKeeper.GetOracles(GetContext(app, true), assetCode)
		require.NoError(t, err)
		require.Len(t, oracles, 2)
	}

	// check the 1st oracle
	{
		oracleObj, err := app.oracleKeeper.GetOracle(GetContext(app, true), assetCode, newOracleAcc1)
		require.NoError(t, err)
		require.True(t, oracleObj.Address.Equals(newOracleAcc1))
	}

	// check the 2nd oracle
	{
		oracleObj, err := app.oracleKeeper.GetOracle(GetContext(app, true), assetCode, newOracleAcc2)
		require.NoError(t, err)
		require.True(t, oracleObj.Address.Equals(newOracleAcc2))
	}
}

func Test_OracleSetOracles(t *testing.T) {
	app, server := newTestDnApp()
	defer app.CloseConnections()
	defer server.Stop()

	genAccs, genAddrs, _, genPrivKeys := CreateGenAccounts(7, GenDefCoins(t))
	_, err := setGenesis(t, app, genAccs)
	require.NoError(t, err)

	nomineeAddr, nomineePrivKey, assetCode := genAddrs[0], genPrivKeys[0], "db2db"

	newOracleAcc1, err := sdk.AccAddressFromHex(secp256k1.GenPrivKey().PubKey().Address().String())
	require.NoError(t, err)
	newOracleAcc2, err := sdk.AccAddressFromHex(secp256k1.GenPrivKey().PubKey().Address().String())
	require.NoError(t, err)
	newOracleAcc3, err := sdk.AccAddressFromHex(secp256k1.GenPrivKey().PubKey().Address().String())
	require.NoError(t, err)
	setOracleAccs := []sdk.AccAddress{newOracleAcc2, newOracleAcc3}

	// set params (add asset / nominees)
	{
		app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{ChainID: chainID, Height: app.LastBlockHeight() + 1}})

		ctx := GetContext(app, false)
		ap := oracle.Params{
			Assets: oracle.Assets{
				oracle.Asset{AssetCode: assetCode, Oracles: oracle.Oracles{}, Active: true},
			},
			Nominees: []string{nomineeAddr.String()},
		}
		app.oracleKeeper.SetParams(ctx, ap)
		assets := app.oracleKeeper.GetAssetParams(ctx)
		require.Equal(t, len(assets), 1)
		require.Equal(t, assets[0].AssetCode, assetCode)

		app.EndBlock(abci.RequestEndBlock{})
		app.Commit()
	}

	// add oracle to the asset
	{
		msg := oracle.MsgAddOracle{
			Oracle:  newOracleAcc1,
			Nominee: nomineeAddr,
			Denom:   assetCode,
		}
		senderAcc := GetAccountCheckTx(app, nomineeAddr)

		tx := genTx([]sdk.Msg{msg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, nomineePrivKey)
		CheckDeliverTx(t, app, tx)
	}

	// check setting oracles to the asset (rewrite the one added before)
	{
		msg := oracle.MsgSetOracles{
			Oracles: oracle.Oracles{{Address: setOracleAccs[0]}, {Address: setOracleAccs[1]}},
			Nominee: nomineeAddr,
			Denom:   assetCode,
		}
		senderAcc := GetAccountCheckTx(app, nomineeAddr)

		tx := genTx([]sdk.Msg{msg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, nomineePrivKey)
		CheckDeliverTx(t, app, tx)

		oracles, err := app.oracleKeeper.GetOracles(GetContext(app, true), assetCode)
		require.NoError(t, err)

		require.Len(t, oracles, len(setOracleAccs))
		for i, acc := range setOracleAccs {
			oracleObj := oracles[i]
			require.True(t, acc.Equals(oracleObj.Address))
		}
	}
}

func Test_OracleAddAsset(t *testing.T) {
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

	// set params (add asset with oracle / nominees)
	{
		app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{ChainID: chainID, Height: app.LastBlockHeight() + 1}})

		ctx := GetContext(app, false)
		ap := oracle.Params{
			Assets: oracle.Assets{
				oracle.Asset{AssetCode: assetCode, Oracles: oracle.Oracles{{Address: newOracleAcc1}}, Active: true},
			},
			Nominees: []string{nomineeAddr.String()},
		}
		app.oracleKeeper.SetParams(ctx, ap)
		assets := app.oracleKeeper.GetAssetParams(ctx)
		require.Equal(t, len(assets), 1)
		require.Equal(t, assets[0].AssetCode, assetCode)
		require.Len(t, assets[0].Oracles, 1)

		app.EndBlock(abci.RequestEndBlock{})
		app.Commit()
	}

	// check adding new asset with existing nominees
	newAssetCode := "dn2test"
	{
		msg := oracle.MsgAddAsset{
			Nominee: nomineeAddr,
			Denom:   newAssetCode,
			Asset:   oracle.NewAsset(newAssetCode, oracle.Oracles{{Address: newOracleAcc1}, {Address: newOracleAcc2}}, true),
		}
		senderAcc := GetAccountCheckTx(app, genAccs[0].Address)

		tx := genTx([]sdk.Msg{msg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, nomineePrivKey)
		CheckDeliverTx(t, app, tx)

		// check the new one added
		newAsset, found := app.oracleKeeper.GetAsset(GetContext(app, true), newAssetCode)
		require.True(t, found)
		require.Equal(t, newAssetCode, newAsset.AssetCode)
		require.Equal(t, true, newAsset.Active)
		require.Len(t, newAsset.Oracles, 2)
		require.True(t, newAsset.Oracles[0].Address.Equals(newOracleAcc1))
		require.True(t, newAsset.Oracles[1].Address.Equals(newOracleAcc2))

		// check the old one still exists
		oldAsset, found := app.oracleKeeper.GetAsset(GetContext(app, true), assetCode)
		require.True(t, found)
		require.Equal(t, assetCode, oldAsset.AssetCode)
		require.Equal(t, true, oldAsset.Active)
		require.Len(t, oldAsset.Oracles, 1)
		require.True(t, oldAsset.Oracles[0].Address.Equals(newOracleAcc1))
	}

	// check adding new asset with non-existing nominees
	{
		msg := oracle.MsgAddOracle{
			Oracle:  newOracleAcc1,
			Nominee: newOracleAcc2,
			Denom:   "non-existing-asset",
		}
		senderAcc := GetAccountCheckTx(app, nomineeAddr)

		tx := genTx([]sdk.Msg{msg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, nomineePrivKey)
		// TODO: add specific error check
		CheckDeliverErrorTx(t, app, tx)
	}
}

func TestOracle_SetAsset(t *testing.T) {
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

	// set params (add asset with oracle / nominees)
	{
		app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{ChainID: chainID, Height: app.LastBlockHeight() + 1}})

		ctx := GetContext(app, false)
		ap := oracle.Params{
			Assets: oracle.Assets{
				oracle.Asset{AssetCode: assetCode, Oracles: oracle.Oracles{{Address: newOracleAcc1}}, Active: true},
			},
			Nominees: []string{nomineeAddr.String()},
		}
		app.oracleKeeper.SetParams(ctx, ap)
		assets := app.oracleKeeper.GetAssetParams(ctx)
		require.Equal(t, len(assets), 1)
		require.Equal(t, assets[0].AssetCode, assetCode)
		require.Len(t, assets[0].Oracles, 1)

		app.EndBlock(abci.RequestEndBlock{})
		app.Commit()
	}

	// check asset created with SetParams exists
	{
		asset, found := app.oracleKeeper.GetAsset(GetContext(app, true), assetCode)
		require.True(t, found)
		require.Equal(t, assetCode, asset.AssetCode)
		require.True(t, asset.Active)
		require.Len(t, asset.Oracles, 1)
		require.Equal(t, newOracleAcc1, asset.Oracles[0].Address)
	}

	// check setting asset (updating)
	{
		updAssetCode := "dn2test1"

		msg := oracle.MsgSetAsset{
			Nominee: nomineeAddr,
			Denom:   assetCode,
			Asset:   oracle.NewAsset(updAssetCode, oracle.Oracles{{Address: newOracleAcc1}, {Address: newOracleAcc2}}, true),
		}
		senderAcc := GetAccountCheckTx(app, nomineeAddr)

		tx := genTx([]sdk.Msg{msg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, nomineePrivKey)
		CheckDeliverTx(t, app, tx)

		asset, found := app.oracleKeeper.GetAsset(GetContext(app, true), updAssetCode)
		require.True(t, found)
		require.Equal(t, updAssetCode, asset.AssetCode)
		require.True(t, asset.Active)
		require.Len(t, asset.Oracles, 2)
		require.True(t, asset.Oracles[0].Address.Equals(newOracleAcc1))
		require.True(t, asset.Oracles[1].Address.Equals(newOracleAcc2))
	}
}

func Test_OraclePostPrices(t *testing.T) {
	app, server := newTestDnApp()
	defer app.CloseConnections()
	defer server.Stop()

	genAccs, genAddrs, _, genPrivKeys := CreateGenAccounts(7, GenDefCoins(t))
	_, err := setGenesis(t, app, genAccs)
	require.NoError(t, err)

	assetCode := "dn2dn"

	// set params (add asset with oracle 0 / nominees)
	{
		nomineeAddr := genAddrs[0]

		app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{ChainID: chainID, Height: app.LastBlockHeight() + 1}})

		ctx := GetContext(app, false)
		ap := oracle.Params{
			Assets: oracle.Assets{
				oracle.Asset{AssetCode: assetCode, Oracles: oracle.Oracles{{Address: genAddrs[0]}}, Active: true},
			},
			Nominees: []string{nomineeAddr.String()},
		}
		app.oracleKeeper.SetParams(ctx, ap)
		assets := app.oracleKeeper.GetAssetParams(ctx)
		require.Equal(t, len(assets), 1)
		require.Equal(t, assets[0].AssetCode, assetCode)

		app.EndBlock(abci.RequestEndBlock{})
		app.Commit()
	}

	// check posting price from non-existing oracle
	{
		senderAcc, senderPrivKey := GetAccountCheckTx(app, genAddrs[1]), genPrivKeys[1]

		msg := oracle.MsgPostPrice{
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

		msg := oracle.MsgPostPrice{
			From:       senderAcc.GetAddress(),
			AssetCode:  "non-existing-asset",
			Price:      sdk.OneInt(),
			ReceivedAt: time.Now(),
		}

		tx := genTx([]sdk.Msg{msg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
		// TODO: add specific error check
		CheckDeliverErrorTx(t, app, tx)
	}

	// set oracles for the asset
	{
		senderAcc, senderPrivKey := GetAccountCheckTx(app, genAddrs[0]), genPrivKeys[0]

		msg := oracle.MsgSetOracles{
			Oracles: oracle.Oracles{{Address: genAddrs[0]}, {Address: genAddrs[1]}, {Address: genAddrs[2]}},
			Nominee: senderAcc.GetAddress(),
			Denom:   assetCode,
		}

		tx := genTx([]sdk.Msg{msg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
		CheckDeliverTx(t, app, tx)
	}

	// check posting price few times from the same oracle
	{
		now := time.Now()
		priceAmount1, priceAmount2 := sdk.NewInt(200000000), sdk.NewInt(100000000)
		priceTimestamp1, priceTimestamp2 := now.Add(1*time.Second), now.Add(2*time.Second)

		// post prices
		app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{ChainID: chainID, Height: app.LastBlockHeight() + 1}})
		{
			senderAcc, senderPrivKey := GetAccount(app, genAddrs[0]), genPrivKeys[0]

			msg := oracle.MsgPostPrice{
				From:       senderAcc.GetAddress(),
				AssetCode:  assetCode,
				Price:      priceAmount1,
				ReceivedAt: priceTimestamp1,
			}

			tx := genTx([]sdk.Msg{msg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
			res := app.Deliver(tx)
			require.True(t, res.IsOK())
		}
		{
			senderAcc, senderPrivKey := GetAccount(app, genAddrs[0]), genPrivKeys[0]

			msg := oracle.MsgPostPrice{
				From:       senderAcc.GetAddress(),
				AssetCode:  assetCode,
				Price:      priceAmount2,
				ReceivedAt: priceTimestamp2,
			}

			tx := genTx([]sdk.Msg{msg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
			res := app.Deliver(tx)
			require.True(t, res.IsOK())
		}
		app.EndBlock(abci.RequestEndBlock{})
		app.Commit()

		// check the last price is the current price
		{
			price := app.oracleKeeper.GetCurrentPrice(GetContext(app, true), assetCode)
			require.True(t, price.Price.Equal(priceAmount2))
			require.True(t, price.ReceivedAt.Equal(priceTimestamp2))
		}

		// check rawPrices
		{
			ctx := GetContext(app, true)
			rawPrices := app.oracleKeeper.GetRawPrices(ctx, assetCode, ctx.BlockHeight()-1)
			require.Len(t, rawPrices, 1)
			require.True(t, priceAmount2.Equal(rawPrices[0].Price))
			require.True(t, priceTimestamp2.Equal(rawPrices[0].ReceivedAt))
			require.Equal(t, assetCode, rawPrices[0].AssetCode)
			require.Equal(t, genAddrs[0], rawPrices[0].OracleAddress)
		}
	}

	// check posting prices from different oracles
	{
		now := time.Now()
		priceValues := []sdk.Int{sdk.NewInt(200000000), sdk.NewInt(100000000), sdk.NewInt(300000000)}
		priceTimestamps := []time.Time{now.Add(1 * time.Second), now.Add(2 * time.Second), now.Add(3 * time.Second)}

		// post prices
		app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{ChainID: chainID, Height: app.LastBlockHeight() + 1}})
		{
			senderAcc, senderPrivKey := GetAccount(app, genAddrs[0]), genPrivKeys[0]

			msg := oracle.MsgPostPrice{
				From:       senderAcc.GetAddress(),
				AssetCode:  assetCode,
				Price:      priceValues[0],
				ReceivedAt: priceTimestamps[0],
			}

			tx := genTx([]sdk.Msg{msg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
			res := app.Deliver(tx)
			require.True(t, res.IsOK())
		}
		{
			senderAcc, senderPrivKey := GetAccount(app, genAddrs[1]), genPrivKeys[1]

			msg := oracle.MsgPostPrice{
				From:       senderAcc.GetAddress(),
				AssetCode:  assetCode,
				Price:      priceValues[1],
				ReceivedAt: priceTimestamps[1],
			}

			tx := genTx([]sdk.Msg{msg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
			res := app.Deliver(tx)
			require.True(t, res.IsOK(), res.Log)
		}
		{
			senderAcc, senderPrivKey := GetAccount(app, genAddrs[2]), genPrivKeys[2]

			msg := oracle.MsgPostPrice{
				From:       senderAcc.GetAddress(),
				AssetCode:  assetCode,
				Price:      priceValues[2],
				ReceivedAt: priceTimestamps[2],
			}

			tx := genTx([]sdk.Msg{msg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
			res := app.Deliver(tx)
			require.True(t, res.IsOK())
		}
		app.EndBlock(abci.RequestEndBlock{})
		app.Commit()

		// check the last price is the median price
		{
			price := app.oracleKeeper.GetCurrentPrice(GetContext(app, true), assetCode)
			require.True(t, price.Price.Equal(priceValues[0]))
			require.True(t, price.ReceivedAt.Equal(priceTimestamps[0]))
		}

		// check rawPrices
		{
			ctx := GetContext(app, true)
			rawPrices := app.oracleKeeper.GetRawPrices(ctx, assetCode, ctx.BlockHeight()-1)
			require.Len(t, rawPrices, 3)
			for i, rawPrice := range rawPrices {
				require.True(t, priceValues[i].Equal(rawPrice.Price))
				require.True(t, priceTimestamps[i].Equal(rawPrice.ReceivedAt))
				require.Equal(t, assetCode, rawPrice.AssetCode)
				require.Equal(t, genAddrs[i], rawPrice.OracleAddress)
			}
		}

		// check rawPrices from the previous block are still exist
		{
			ctx := GetContext(app, true)
			rawPrices := app.oracleKeeper.GetRawPrices(ctx, assetCode, ctx.BlockHeight()-2)
			require.Len(t, rawPrices, 1)
		}
	}
}
