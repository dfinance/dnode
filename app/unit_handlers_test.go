// +build integ

package app

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/dfinance/dnode/cmd/config/genesis/defaults"
	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/common_vm"
	"github.com/dfinance/dnode/x/currencies"
	"github.com/dfinance/dnode/x/markets"
	"github.com/dfinance/dnode/x/oracle"
	"github.com/dfinance/dnode/x/orders"
	"github.com/dfinance/dnode/x/vm"
	"github.com/dfinance/dnode/x/vm/client/vm_client"
	"github.com/dfinance/dvm-proto/go/compiler_grpc"
)

func TestHandlers_CheckEvents(t *testing.T) {
	t.Parallel()

	app, dvmAddress, appStop := NewTestDnAppDVM(t, log.AllowInfoWith("module", "x/vm"))
	defer appStop()

	genAccs, _, _, genPrivKeys := CreateGenAccounts(7, GenDefCoins(t))
	CheckSetGenesisDVM(t, app, genAccs)

	client1Addr := genAccs[0].Address
	client1LibraAddr := common_vm.Bech32ToLibra(client1Addr)

	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{ChainID: chainID, Height: app.LastBlockHeight() + 1}})
	app.EndBlock(abci.RequestEndBlock{})
	app.Commit()

	equalEvents := func(t *testing.T, logs string, expectedLen int) {
		logList := sdk.ABCIMessageLogs{}
		err := json.Unmarshal([]byte(logs), &logList)
		require.NoError(t, err)
		require.Len(t, logList, expectedLen)

		var prevLog []byte
		for _, logItem := range logList {
			msgLog, err := json.Marshal(logItem.Events)
			require.NoError(t, err)
			if prevLog != nil {
				require.Equal(t, prevLog, msgLog)
			}
			prevLog = msgLog
		}
	}

	similarAttributes := func(t *testing.T, logs string, expectedLen int) {
		logList := sdk.ABCIMessageLogs{}
		err := json.Unmarshal([]byte(logs), &logList)
		require.NoError(t, err)
		require.Len(t, logList, expectedLen)

		var prevLog sdk.ABCIMessageLog
		isFirst := true
		for _, logItem := range logList {
			if !isFirst {
				for iEv, ev := range logItem.Events {
					require.Equal(t, prevLog.Events[iEv].Type, ev.Type)
					require.Equal(t, len(prevLog.Events[iEv].Attributes), len(ev.Attributes),
						fmt.Sprintf("Event: %d", iEv))

					for iAttr, attr := range ev.Attributes {
						require.Equal(t, prevLog.Events[iEv].Attributes[iAttr].Key, attr.Key,
							fmt.Sprintf("Event: %d, Attribute: %d", iEv, iAttr))
					}
				}
			}
			isFirst = false
			prevLog = logItem
		}
	}

	// Check vm execute script handler
	{
		script := `
			script {
				use 0x1::Account;
				use 0x1::XFI;
	
				fun main(account: &signer, amount: u128) {
					let xfi = Account::withdraw_from_sender<XFI::T>(account, amount);
					Account::deposit_to_sender<XFI::T>(account, xfi);
				}
			}
		`
		arg1, aErr1 := vm_client.NewU128ScriptArg("100")
		require.NoError(t, aErr1)

		senderAcc, senderPrivKey := GetAccountCheckTx(app, client1Addr), genPrivKeys[0]

		byteCode, compileErr := vm_client.Compile(dvmAddress, &compiler_grpc.SourceFiles{
			Units: []*compiler_grpc.CompilationUnit{
				{
					Text: script,
					Name: "script",
				},
			},
			Address: client1LibraAddr,
		})
		require.NoError(t, compileErr)
		require.Len(t, byteCode, 1)

		msg := vm.MsgExecuteScript{
			Signer: client1Addr,
			Script: byteCode[0].ByteCode,
			Args:   []vm.ScriptArg{arg1},
		}

		txMsg := []sdk.Msg{msg, msg, msg}
		tx := GenTx(txMsg, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
		app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{ChainID: chainID, Height: app.LastBlockHeight() + 1}})

		_, res, err := app.Deliver(tx)
		require.NoError(t, err, res)

		equalEvents(t, res.Log, len(txMsg))

		app.EndBlock(abci.RequestEndBlock{})
		app.Commit()
	}

	// Check vm deploy module handler
	{
		module := `
			address 0x1 {
				module Foo {
					resource struct U64 {val: u64}
					public fun store_u64(sender: &signer) {
						let value = U64 {val: 1};
						move_to<U64>(sender, value);
					}
				}
	
				module Bar {
					public fun sub(a: u64, b: u64): u64 {
						a - b
					}
				}
			}
		`

		senderAcc, senderPrivKey := GetAccountCheckTx(app, client1Addr), genPrivKeys[0]

		byteCode, compileErr := vm_client.Compile(dvmAddress, &compiler_grpc.SourceFiles{
			Units: []*compiler_grpc.CompilationUnit{
				{
					Text: module,
					Name: "module",
				},
			},
			Address: client1LibraAddr,
		})
		require.NoError(t, compileErr)
		require.Len(t, byteCode, 2)

		msg := vm.MsgDeployModule{
			Signer: client1Addr,
			Module: []vm.Contract{byteCode[0].ByteCode},
		}

		txMsg := []sdk.Msg{msg, msg, msg}
		tx := GenTx(txMsg, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
		app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{ChainID: chainID, Height: app.LastBlockHeight() + 1}})

		_, res, err := app.Deliver(tx)
		require.NoError(t, err, res)

		equalEvents(t, res.Log, len(txMsg))

		app.EndBlock(abci.RequestEndBlock{})
		app.Commit()
	}

	// Check market CreateMarket handler
	{
		senderAcc, senderPrivKey := GetAccountCheckTx(app, client1Addr), genPrivKeys[0]

		msgBtcEth := markets.MsgCreateMarket{
			From:            client1Addr,
			BaseAssetDenom:  "btc",
			QuoteAssetDenom: "eth",
		}

		msgBtcUsdt := markets.MsgCreateMarket{
			From:            client1Addr,
			BaseAssetDenom:  "xfi",
			QuoteAssetDenom: "btc",
		}

		txMsg := []sdk.Msg{msgBtcEth, msgBtcUsdt}
		tx := GenTx(txMsg, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
		app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{ChainID: chainID, Height: app.LastBlockHeight() + 1}})

		_, res, err := app.Deliver(tx)
		require.NoError(t, err, res)

		similarAttributes(t, res.Log, len(txMsg))

		app.EndBlock(abci.RequestEndBlock{})
		app.Commit()
	}

	// Check orders PostOrder handler
	{
		senderAcc, senderPrivKey := GetAccountCheckTx(app, client1Addr), genPrivKeys[0]

		msg := orders.MsgPostOrder{
			Owner:     client1Addr,
			AssetCode: dnTypes.AssetCode("xfi_btc"),
			Direction: orders.AskDirection,
			Price:     sdk.NewUint(1),
			Quantity:  sdk.NewUint(1),
			TtlInSec:  1000,
		}

		txMsg := []sdk.Msg{msg, msg}
		tx := GenTx(txMsg, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
		app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{ChainID: chainID, Height: app.LastBlockHeight() + 1}})

		_, res, err := app.Deliver(tx)
		require.NoError(t, err, res)

		similarAttributes(t, res.Log, len(txMsg))

		app.EndBlock(abci.RequestEndBlock{})
		app.Commit()
	}

	// Check oracle AddOracle handler
	{
		senderAcc, senderPrivKey := GetAccountCheckTx(app, client1Addr), genPrivKeys[0]
		ac := dnTypes.AssetCode("xfi_btc")

		ap := oracle.Params{
			Assets: oracle.Assets{
				oracle.Asset{AssetCode: ac, Oracles: oracle.Oracles{}, Active: true},
			},
			Nominees: []string{client1Addr.String()},
		}

		app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{ChainID: chainID, Height: app.LastBlockHeight() + 1}})

		ctx := GetContext(app, false)
		app.oracleKeeper.SetParams(ctx, ap)

		msg := oracle.MsgAddOracle{
			Nominee:   client1Addr,
			Oracle:    client1Addr,
			AssetCode: ac,
		}

		msg2 := oracle.MsgAddOracle{
			Nominee:   client1Addr,
			Oracle:    genAccs[1].Address,
			AssetCode: ac,
		}

		txMsg := []sdk.Msg{msg, msg2}
		tx := GenTx(txMsg, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)

		_, res, err := app.Deliver(tx)
		require.NoError(t, err, res)

		similarAttributes(t, res.Log, len(txMsg))

		app.EndBlock(abci.RequestEndBlock{})
		app.Commit()
	}

	// Check oracle PostPrice handler
	{
		senderAcc, senderPrivKey := GetAccountCheckTx(app, client1Addr), genPrivKeys[0]
		ac := dnTypes.AssetCode("xfi_btc")
		app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{ChainID: chainID, Height: app.LastBlockHeight() + 1}})

		msg := oracle.MsgPostPrice{
			ReceivedAt: time.Now(),
			From:       client1Addr,
			AssetCode:  ac,
			AskPrice:   sdk.NewInt(1000),
			BidPrice:   sdk.NewInt(995),
		}

		txMsg := []sdk.Msg{msg, msg}
		tx := GenTx(txMsg, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)

		_, res, err := app.Deliver(tx)
		require.NoError(t, err, res)

		equalEvents(t, res.Log, len(txMsg))

		app.EndBlock(abci.RequestEndBlock{})
		app.Commit()
	}

	// Check currencies WithdrawCurrency handler
	{
		senderAcc, senderPrivKey := GetAccountCheckTx(app, client1Addr), genPrivKeys[0]
		app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{ChainID: chainID, Height: app.LastBlockHeight() + 1}})

		msg := currencies.MsgWithdrawCurrency{
			Coin:           sdk.NewCoin(defaults.MainDenom, sdk.NewInt(1)),
			Payer:          client1Addr,
			PegZonePayee:   "pegZone",
			PegZoneChainID: "peg-chain",
		}

		txMsg := []sdk.Msg{msg, msg}
		tx := GenTx(txMsg, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)

		_, res, err := app.Deliver(tx)
		require.NoError(t, err, res)

		similarAttributes(t, res.Log, len(txMsg))

		app.EndBlock(abci.RequestEndBlock{})
		app.Commit()
	}
}
