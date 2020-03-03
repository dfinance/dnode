package app

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	"github.com/WingsDao/wings-blockchain/x/oracle"
)

func Test_AddOracle(t *testing.T) {
	app, server := newTestWbApp()
	defer app.CloseConnections()
	defer server.Stop()

	genCoins, err := sdk.ParseCoins("1000000000000000wings")
	require.NoError(t, err)
	genAccs, addrs, _, privKeys := CreateGenAccounts(7, genCoins)
	_, err = setGenesis(t, app, genAccs)
	require.NoError(t, err)

	newOracleAcc1, err := sdk.AccAddressFromHex(secp256k1.GenPrivKey().PubKey().Address().String())
	require.NoError(t, err)
	newOracleAcc2, err := sdk.AccAddressFromHex(secp256k1.GenPrivKey().PubKey().Address().String())
	require.NoError(t, err)
	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{ChainID: chainID, Height: app.LastBlockHeight() + 1}})
	{
		ctx := GetContext(app, false)
		ap := oracle.Params{
			Assets: oracle.Assets{
				oracle.Asset{AssetCode: "wb2wb", Oracles: oracle.Oracles{}, Active: true},
			},
			Nominees: []string{addrs[0].String()},
		}
		app.oracleKeeper.SetParams(ctx, ap)
		assets := app.oracleKeeper.GetAssetParams(ctx)
		require.Equal(t, len(assets), 1)
		require.Equal(t, assets[0].AssetCode, "wb2wb")
	}
	app.EndBlock(abci.RequestEndBlock{})
	app.Commit()

	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{ChainID: chainID, Height: app.LastBlockHeight() + 1}})
	{
		msg := oracle.MsgAddOracle{
			Oracle:  newOracleAcc1,
			Nominee: addrs[0],
			Denom:   "wb2wb",
		}
		acc := GetAccountCheckTx(app, genAccs[0].Address)
		tx := genTx([]sdk.Msg{msg}, []uint64{acc.GetAccountNumber()}, []uint64{acc.GetSequence()}, privKeys[0])

		res := app.Deliver(tx)
		require.True(t, res.IsOK())
	}
	app.EndBlock(abci.RequestEndBlock{})
	app.Commit()

	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{ChainID: chainID, Height: app.LastBlockHeight() + 1}})
	{
		msg := oracle.MsgAddOracle{
			Oracle:  newOracleAcc2,
			Nominee: addrs[0],
			Denom:   "wb2wb",
		}
		acc := GetAccountCheckTx(app, genAccs[0].Address)
		tx := genTx([]sdk.Msg{msg}, []uint64{acc.GetAccountNumber()}, []uint64{acc.GetSequence()}, privKeys[0])

		res := app.Deliver(tx)
		require.True(t, res.IsOK())
	}
	app.EndBlock(abci.RequestEndBlock{})
	app.Commit()

	oracles, err := app.oracleKeeper.GetOracles(GetContext(app, true), "wb2wb")
	require.NoError(t, err)
	require.Equal(t, int(2), len(oracles))
	oracle, err := app.oracleKeeper.GetOracle(GetContext(app, true), "wb2wb", newOracleAcc1)
	require.NoError(t, err)
	require.True(t, oracle.Address.Equals(newOracleAcc1))

	oracle, err = app.oracleKeeper.GetOracle(GetContext(app, true), "wb2wb", newOracleAcc2)
	require.NoError(t, err)
	require.True(t, oracle.Address.Equals(newOracleAcc2))
}

func Test_SetOracles(t *testing.T) {
	app, server := newTestWbApp()
	defer app.CloseConnections()
	defer server.Stop()

	genCoins, err := sdk.ParseCoins("1000000000000000wings")
	require.NoError(t, err)
	genAccs, addrs, _, privKeys := CreateGenAccounts(7, genCoins)
	_, err = setGenesis(t, app, genAccs)
	require.NoError(t, err)

	newOracleAcc1, err := sdk.AccAddressFromHex(secp256k1.GenPrivKey().PubKey().Address().String())
	require.NoError(t, err)
	newOracleAcc2, err := sdk.AccAddressFromHex(secp256k1.GenPrivKey().PubKey().Address().String())
	require.NoError(t, err)
	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{ChainID: chainID, Height: app.LastBlockHeight() + 1}})
	{
		ctx := GetContext(app, false)
		ap := oracle.Params{
			Assets: oracle.Assets{
				oracle.Asset{AssetCode: "wb2wb", Oracles: oracle.Oracles{}, Active: true},
			},
			Nominees: []string{addrs[0].String()},
		}
		app.oracleKeeper.SetParams(ctx, ap)
		assets := app.oracleKeeper.GetAssetParams(ctx)
		require.Equal(t, len(assets), 1)
		require.Equal(t, assets[0].AssetCode, "wb2wb")
	}
	app.EndBlock(abci.RequestEndBlock{})
	app.Commit()

	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{ChainID: chainID, Height: app.LastBlockHeight() + 1}})
	{
		msg := oracle.MsgAddOracle{
			Oracle:  newOracleAcc1,
			Nominee: addrs[0],
			Denom:   "wb2wb",
		}
		acc := GetAccountCheckTx(app, genAccs[0].Address)
		tx := genTx([]sdk.Msg{msg}, []uint64{acc.GetAccountNumber()}, []uint64{acc.GetSequence()}, privKeys[0])

		res := app.Deliver(tx)
		require.True(t, res.IsOK())
	}
	app.EndBlock(abci.RequestEndBlock{})
	app.Commit()

	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{ChainID: chainID, Height: app.LastBlockHeight() + 1}})
	{
		ctx := GetContext(app, false)
		require.NoError(t, err)
		msg := oracle.MsgSetOracles{
			Oracles: oracle.Oracles{{Address: newOracleAcc2}},
			Nominee: addrs[0],
			Denom:   "wb2wb",
		}
		acc := GetAccountCheckTx(app, genAccs[0].Address)
		tx := genTx([]sdk.Msg{msg}, []uint64{acc.GetAccountNumber()}, []uint64{acc.GetSequence()}, privKeys[0])

		res := app.Deliver(tx)
		require.True(t, res.IsOK())
		oracles, err := app.oracleKeeper.GetOracles(ctx, "wb2wb")
		require.NoError(t, err)
		require.Equal(t, int(1), len(oracles))
		require.True(t, oracles[0].Address.Equals(newOracleAcc2))
	}
	app.EndBlock(abci.RequestEndBlock{})
	app.Commit()
}

func Test_AddAsset(t *testing.T) {
	app, server := newTestWbApp()
	defer app.CloseConnections()
	defer server.Stop()

	genCoins, err := sdk.ParseCoins("1000000000000000wings")
	require.NoError(t, err)
	genAccs, addrs, _, privKeys := CreateGenAccounts(7, genCoins)
	_, err = setGenesis(t, app, genAccs)
	require.NoError(t, err)

	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{ChainID: chainID, Height: app.LastBlockHeight() + 1}})
	{
		ctx := GetContext(app, false)
		ap := oracle.Params{
			Assets: oracle.Assets{
				oracle.Asset{AssetCode: "wb2wb", Oracles: oracle.Oracles{}, Active: true},
			},
			Nominees: []string{addrs[0].String()},
		}
		app.oracleKeeper.SetParams(ctx, ap)
		assets := app.oracleKeeper.GetAssetParams(ctx)
		require.Equal(t, len(assets), 1)
		require.Equal(t, assets[0].AssetCode, "wb2wb")
	}
	app.EndBlock(abci.RequestEndBlock{})
	app.Commit()

	newOracleAcc1, err := sdk.AccAddressFromHex(secp256k1.GenPrivKey().PubKey().Address().String())
	require.NoError(t, err)
	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{ChainID: chainID, Height: app.LastBlockHeight() + 1}})
	{
		msg := oracle.MsgAddOracle{
			Oracle:  newOracleAcc1,
			Nominee: addrs[0],
			Denom:   "wb2wb",
		}
		acc := GetAccountCheckTx(app, genAccs[0].Address)
		tx := genTx([]sdk.Msg{msg}, []uint64{acc.GetAccountNumber()}, []uint64{acc.GetSequence()}, privKeys[0])

		res := app.Deliver(tx)
		require.True(t, res.IsOK())
	}
	app.EndBlock(abci.RequestEndBlock{})
	app.Commit()

	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{ChainID: chainID, Height: app.LastBlockHeight() + 1}})
	{
		msg := oracle.MsgAddAsset{
			Nominee: addrs[0],
			Denom:   "wb2test",
			Asset:   oracle.NewAsset("wb2test", oracle.Oracles{{Address: newOracleAcc1}}, true),
		}
		acc := GetAccountCheckTx(app, genAccs[0].Address)
		tx := genTx([]sdk.Msg{msg}, []uint64{acc.GetAccountNumber()}, []uint64{acc.GetSequence()}, privKeys[0])

		res := app.Deliver(tx)
		require.True(t, res.IsOK())

		ctx := GetContext(app, false)
		asset, found := app.oracleKeeper.GetAsset(ctx, "wb2test")
		require.True(t, found)
		require.Equal(t, "wb2test", asset.AssetCode)
		require.True(t, asset.Oracles[0].Address.Equals(newOracleAcc1))
	}
	app.EndBlock(abci.RequestEndBlock{})
	app.Commit()
}

func Test_SetAsset(t *testing.T) {
	app, server := newTestWbApp()
	defer app.CloseConnections()
	defer server.Stop()

	genCoins, err := sdk.ParseCoins("1000000000000000wings")
	require.NoError(t, err)
	genAccs, addrs, _, privKeys := CreateGenAccounts(7, genCoins)
	_, err = setGenesis(t, app, genAccs)
	require.NoError(t, err)

	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{ChainID: chainID, Height: app.LastBlockHeight() + 1}})
	{
		ctx := GetContext(app, false)
		ap := oracle.Params{
			Assets: oracle.Assets{
				oracle.Asset{AssetCode: "wb2wb", Oracles: oracle.Oracles{}, Active: true},
			},
			Nominees: []string{addrs[0].String()},
		}
		app.oracleKeeper.SetParams(ctx, ap)
		assets := app.oracleKeeper.GetAssetParams(ctx)
		require.Equal(t, len(assets), 1)
		require.Equal(t, assets[0].AssetCode, "wb2wb")
	}
	app.EndBlock(abci.RequestEndBlock{})
	app.Commit()

	newOracleAcc1, err := sdk.AccAddressFromHex(secp256k1.GenPrivKey().PubKey().Address().String())
	require.NoError(t, err)
	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{ChainID: chainID, Height: app.LastBlockHeight() + 1}})
	{
		msg := oracle.MsgAddOracle{
			Oracle:  newOracleAcc1,
			Nominee: addrs[0],
			Denom:   "wb2wb",
		}
		acc := GetAccountCheckTx(app, genAccs[0].Address)
		tx := genTx([]sdk.Msg{msg}, []uint64{acc.GetAccountNumber()}, []uint64{acc.GetSequence()}, privKeys[0])

		res := app.Deliver(tx)
		require.True(t, res.IsOK())
	}
	app.EndBlock(abci.RequestEndBlock{})
	app.Commit()

	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{ChainID: chainID, Height: app.LastBlockHeight() + 1}})
	{
		msg := oracle.MsgAddAsset{
			Nominee: addrs[0],
			Denom:   "wb2test",
			Asset:   oracle.NewAsset("wb2test", oracle.Oracles{{Address: newOracleAcc1}}, true),
		}
		acc := GetAccountCheckTx(app, genAccs[0].Address)
		tx := genTx([]sdk.Msg{msg}, []uint64{acc.GetAccountNumber()}, []uint64{acc.GetSequence()}, privKeys[0])

		res := app.Deliver(tx)
		require.True(t, res.IsOK())

		ctx := GetContext(app, false)
		asset, found := app.oracleKeeper.GetAsset(ctx, "wb2test")
		require.True(t, found)
		require.Equal(t, "wb2test", asset.AssetCode)
		require.True(t, asset.Oracles[0].Address.Equals(newOracleAcc1))
	}
	app.EndBlock(abci.RequestEndBlock{})
	app.Commit()

	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{ChainID: chainID, Height: app.LastBlockHeight() + 1}})
	{
		msg := oracle.MsgSetAsset{
			Nominee: addrs[0],
			Denom:   "wb2test",
			Asset:   oracle.NewAsset("wb2test1", oracle.Oracles{{Address: newOracleAcc1}}, true),
		}
		acc := GetAccountCheckTx(app, genAccs[0].Address)
		tx := genTx([]sdk.Msg{msg}, []uint64{acc.GetAccountNumber()}, []uint64{acc.GetSequence()}, privKeys[0])

		res := app.Deliver(tx)
		require.True(t, res.IsOK())

		ctx := GetContext(app, false)
		asset, found := app.oracleKeeper.GetAsset(ctx, "wb2test1")
		require.True(t, found)
		require.Equal(t, "wb2test1", asset.AssetCode)
		require.True(t, asset.Oracles[0].Address.Equals(newOracleAcc1))
	}
	app.EndBlock(abci.RequestEndBlock{})
	app.Commit()
}

func Test_SetPostPrice(t *testing.T) {
	app, server := newTestWbApp()
	defer app.CloseConnections()
	defer server.Stop()

	genCoins, err := sdk.ParseCoins("1000000000000000wings")
	require.NoError(t, err)
	genAccs, addrs, _, privKeys := CreateGenAccounts(7, genCoins)
	_, err = setGenesis(t, app, genAccs)
	require.NoError(t, err)

	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{ChainID: chainID, Height: app.LastBlockHeight() + 1}})
	{
		ctx := GetContext(app, false)
		ap := oracle.Params{
			Assets: oracle.Assets{
				oracle.Asset{AssetCode: "wb2wb", Oracles: oracle.Oracles{}, Active: true},
			},
			Nominees: []string{addrs[0].String()},
		}
		app.oracleKeeper.SetParams(ctx, ap)
		assets := app.oracleKeeper.GetAssetParams(ctx)
		require.Equal(t, len(assets), 1)
		require.Equal(t, assets[0].AssetCode, "wb2wb")
	}
	app.EndBlock(abci.RequestEndBlock{})
	app.Commit()

	require.NoError(t, err)
	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{ChainID: chainID, Height: app.LastBlockHeight() + 1}})
	{
		msg := oracle.MsgAddOracle{
			Oracle:  addrs[1],
			Nominee: addrs[0],
			Denom:   "wb2wb",
		}
		acc := GetAccount(app, genAccs[0].Address)
		tx := genTx([]sdk.Msg{msg}, []uint64{acc.GetAccountNumber()}, []uint64{acc.GetSequence()}, privKeys[0])

		res := app.Deliver(tx)
		require.True(t, res.IsOK())
	}
	app.EndBlock(abci.RequestEndBlock{})
	app.Commit()

	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{ChainID: chainID, Height: app.LastBlockHeight() + 1}})
	{
		price := sdk.NewInt(100000000)
		msg := oracle.MsgPostPrice{
			From:       addrs[1],
			AssetCode:  "wb2wb",
			Price:      price,
			ReceivedAt: time.Now(),
		}
		acc := GetAccount(app, genAccs[1].Address)
		tx := genTx([]sdk.Msg{msg}, []uint64{acc.GetAccountNumber()}, []uint64{acc.GetSequence()}, privKeys[1])

		res := app.Deliver(tx)
		require.True(t, res.IsOK())
	}

	{
		price := sdk.NewInt(200000000)
		msg := oracle.MsgPostPrice{
			From:       addrs[1],
			AssetCode:  "wb2wb",
			Price:      price,
			ReceivedAt: time.Now(),
		}
		acc := GetAccount(app, genAccs[1].Address)
		tx := genTx([]sdk.Msg{msg}, []uint64{acc.GetAccountNumber()}, []uint64{acc.GetSequence()}, privKeys[1])

		res := app.Deliver(tx)
		require.True(t, res.IsOK())
	}

	{
		price := sdk.NewInt(300000000)
		msg := oracle.MsgPostPrice{
			From:       addrs[1],
			AssetCode:  "wb2wb",
			Price:      price,
			ReceivedAt: time.Now(),
		}
		acc := GetAccount(app, genAccs[1].Address)
		tx := genTx([]sdk.Msg{msg}, []uint64{acc.GetAccountNumber()}, []uint64{acc.GetSequence()}, privKeys[1])

		res := app.Deliver(tx)
		require.True(t, res.IsOK())
	}
	app.EndBlock(abci.RequestEndBlock{})
	app.Commit()

	price := app.oracleKeeper.GetCurrentPrice(GetContext(app, true), "wb2wb")
	require.True(t, price.Price.Equal(sdk.NewInt(300000000)))

}
