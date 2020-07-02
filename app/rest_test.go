// +build rest

package app

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"

	cliTester "github.com/dfinance/dnode/helpers/tests/clitester"
	dnTypes "github.com/dfinance/dnode/helpers/types"
	ccTypes "github.com/dfinance/dnode/x/currencies/types"
	marketTypes "github.com/dfinance/dnode/x/markets"
	msTypes "github.com/dfinance/dnode/x/multisig/types"
	"github.com/dfinance/dnode/x/oracle"
	orderTypes "github.com/dfinance/dnode/x/orders"
	"github.com/dfinance/dnode/x/vm"
)

func Test_CurrencyRest(t *testing.T) {
	t.Parallel()
	ct := cliTester.New(t, false)
	defer ct.Close()
	ct.StartRestServer(false)

	recipientAddr := ct.Accounts["validator1"].Address
	curAmount, curDecimals, denom, issueId := sdk.NewInt(100), int8(0), "btc", "issue1"
	destroyAmounts := make([]sdk.Int, 0)

	// issue currency
	ct.TxCurrenciesIssue(recipientAddr, recipientAddr, denom, curAmount, curDecimals, issueId).CheckSucceeded()
	ct.ConfirmCall(issueId)

	// check getIssue endpoint
	{
		req, respMsg := ct.RestQueryCurrenciesIssue(issueId)
		req.CheckSucceeded()

		require.Equal(t, denom, respMsg.Symbol)
		require.True(t, respMsg.Amount.Equal(curAmount))
		require.Equal(t, recipientAddr, respMsg.Recipient.String())

		// check incorrect inputs
		{
			// non-existing issueID
			{
				req, _ := ct.RestQueryCurrenciesIssue("non_existing_ID")
				req.CheckFailed(http.StatusInternalServerError, ccTypes.ErrWrongIssueID)
			}
		}
	}

	// check getCurrency endpoint
	{
		req, respMsg := ct.RestQueryCurrenciesCurrency(denom)
		req.CheckSucceeded()

		require.Equal(t, denom, respMsg.Symbol)
		require.True(t, respMsg.Supply.Equal(curAmount))
		require.Equal(t, curDecimals, respMsg.Decimals)

		// check incorrect inputs
		{
			// non-existing symbol
			{
				req, _ := ct.RestQueryCurrenciesCurrency("non_existing_symbol")
				req.CheckFailed(http.StatusInternalServerError, ccTypes.ErrNotExistCurrency)
			}
		}
	}

	// check getDestroys endpoint (no destroys)
	{
		req, respMsg := ct.RestQueryCurrenciesDestroys(1, nil)
		req.CheckSucceeded()

		require.Len(t, *respMsg, 0)
	}

	// destroy currency
	newAmount := sdk.NewInt(50)
	curAmount = curAmount.Sub(newAmount)
	ct.TxCurrenciesDestroy(recipientAddr, recipientAddr, denom, newAmount).CheckSucceeded()
	destroyAmounts = append(destroyAmounts, newAmount)

	// check getDestroy endpoint
	{
		req, respMsg := ct.RestQueryCurrenciesDestroy(sdk.NewInt(0))
		req.CheckSucceeded()

		require.Equal(t, int64(0), respMsg.ID.Int64())
		require.Equal(t, ct.IDs.ChainID, respMsg.ChainID)
		require.Equal(t, denom, respMsg.Symbol)
		require.True(t, respMsg.Amount.Equal(newAmount))
		require.Equal(t, recipientAddr, respMsg.Spender.String())
		require.Equal(t, recipientAddr, respMsg.Recipient)

		// check incorrect inputs
		{
			// invalid destroyID
			{
				req, _ := ct.RestQueryCurrenciesDestroy(sdk.NewInt(0))
				req.ModifySubPath("0", "abc")
				req.CheckFailed(http.StatusInternalServerError, nil)
			}

			// non-existing destroyID
			{
				req, respMsg := ct.RestQueryCurrenciesDestroy(sdk.NewInt(1))
				req.CheckSucceeded()

				require.Empty(t, respMsg.ChainID)
				require.Empty(t, respMsg.Symbol)
				require.True(t, respMsg.Amount.IsZero())
			}
		}
	}

	// destroy currency once more
	newAmount = sdk.NewInt(25)
	curAmount = curAmount.Sub(newAmount)
	ct.TxCurrenciesDestroy(recipientAddr, recipientAddr, denom, newAmount).CheckSucceeded()
	destroyAmounts = append(destroyAmounts, newAmount)

	// check getDestroys endpoint
	{
		req, respMsg := ct.RestQueryCurrenciesDestroys(1, nil)
		req.CheckSucceeded()

		require.Len(t, *respMsg, len(destroyAmounts))
		for i, amount := range destroyAmounts {
			destroy := (*respMsg)[i]
			require.Equal(t, int64(i), destroy.ID.Int64())
			require.Equal(t, ct.IDs.ChainID, destroy.ChainID)
			require.Equal(t, denom, destroy.Symbol)
			require.True(t, destroy.Amount.Equal(amount))
			require.Equal(t, recipientAddr, destroy.Spender.String())
			require.Equal(t, recipientAddr, destroy.Recipient)
		}

		// check incorrect inputs
		{
			// invalid "page" value
			{
				req, _ := ct.RestQueryCurrenciesDestroys(1, nil)
				req.ModifySubPath("1", "abc")
				req.CheckFailed(http.StatusInternalServerError, nil)
			}

			// invalid "limit" value
			{
				limit := 1
				req, _ := ct.RestQueryCurrenciesDestroys(1, &limit)
				req.ModifyUrlValues("limit", "abc")
				req.CheckFailed(http.StatusInternalServerError, nil)
			}
		}
	}
}

func Test_MSRest(t *testing.T) {
	t.Parallel()
	ct := cliTester.New(t, false)
	defer ct.Close()
	ct.StartRestServer(false)

	senderAddr := ct.Accounts["validator1"].Address
	msgIDs := make([]string, 0)

	// submit remove validator call (1st one)
	{
		targetValidator := ct.Accounts["validator3"].Address
		msgID := fmt.Sprintf("removeValidator:%s", targetValidator)
		msgIDs = append(msgIDs, msgID)

		ct.TxPoaRemoveValidator(senderAddr, targetValidator, msgID).CheckSucceeded()
	}

	// submit remove validator call (2nd one)
	{
		targetValidator := ct.Accounts["validator2"].Address
		msgID := fmt.Sprintf("removeValidator:%s", targetValidator)
		msgIDs = append(msgIDs, msgID)

		ct.TxPoaRemoveValidator(senderAddr, targetValidator, msgID).CheckSucceeded()
	}

	// check getCalls endpoint
	{
		req, respMsg := ct.RestQueryMultiSigCalls()
		req.CheckSucceeded()

		require.Len(t, *respMsg, 2)
		for i, call := range *respMsg {
			require.Len(t, call.Votes, 1)
			require.Equal(t, senderAddr, call.Votes[0].String())
			require.Equal(t, uint64(i), call.Call.MsgID)
			require.Equal(t, senderAddr, call.Call.Creator.String())
			require.Equal(t, msgIDs[i], call.Call.UniqueID)
		}
	}

	// check getCall endpoint
	{
		req, respMsg := ct.RestQueryMultiSigCall(0)
		req.CheckSucceeded()

		require.Len(t, respMsg.Votes, 1)
		require.Equal(t, senderAddr, respMsg.Votes[0].String())
		require.Equal(t, uint64(0), respMsg.Call.MsgID)
		require.Equal(t, senderAddr, respMsg.Call.Creator.String())
		require.Equal(t, msgIDs[0], respMsg.Call.UniqueID)

		// check invalid inputs
		{
			// invalid "id"
			{
				req, _ := ct.RestQueryMultiSigCall(0)
				req.ModifySubPath("0", "-1")
				req.CheckFailed(http.StatusInternalServerError, nil)
			}

			// non-existing "id"
			{
				req, _ := ct.RestQueryMultiSigCall(2)
				req.CheckFailed(http.StatusInternalServerError, msTypes.ErrWrongCallId)
			}
		}
	}

	// check getCallByUnique endpoint
	{
		req, respMsg := ct.RestQueryMultiSigUnique(msgIDs[0])
		req.CheckSucceeded()

		require.Len(t, respMsg.Votes, 1)
		require.Equal(t, senderAddr, respMsg.Votes[0].String())
		require.Equal(t, uint64(0), respMsg.Call.MsgID)
		require.Equal(t, senderAddr, respMsg.Call.Creator.String())
		require.Equal(t, msgIDs[0], respMsg.Call.UniqueID)

		// check invalid inputs
		{
			// non-existing "unique"
			{
				req, _ := ct.RestQueryMultiSigUnique("non-existing-UNIQUE")
				req.CheckFailed(http.StatusInternalServerError, msTypes.ErrNotFoundUniqueID)
			}
		}
	}
}

func Test_OracleRest(t *testing.T) {
	t.Parallel()
	ct := cliTester.New(t, false)
	defer ct.Close()
	ct.StartRestServer(false)

	oracleName1, oracleName2 := "oracle1", "oracle2"
	oracleAddr1, oracleAddr2 := ct.Accounts[oracleName1].Address, ct.Accounts[oracleName2].Address

	// check getAssets endpoint
	{
		req, respMsg := ct.RestQueryOracleAssets()
		req.CheckSucceeded()

		require.Len(t, *respMsg, 1)
		asset := (*respMsg)[0]
		require.Equal(t, ct.DefAssetCode, asset.AssetCode)
		require.Len(t, asset.Oracles, 2)
		require.Equal(t, oracleAddr1, asset.Oracles[0].Address.String())
		require.Equal(t, oracleAddr2, asset.Oracles[1].Address.String())
		require.True(t, asset.Active)
	}

	now := time.Now()
	postPrices := []struct {
		AssetCode     string
		SenderIdx     uint
		OracleName    string
		OracleAddress string
		Price         sdk.Int
		ReceivedAt    time.Time
		BlockHeight   int64
	}{
		{
			AssetCode:     ct.DefAssetCode,
			SenderIdx:     0,
			OracleName:    oracleName1,
			OracleAddress: oracleAddr1,
			Price:         sdk.NewInt(100),
			ReceivedAt:    now,
			BlockHeight:   0,
		},
		{
			AssetCode:     ct.DefAssetCode,
			SenderIdx:     1,
			OracleName:    oracleName2,
			OracleAddress: oracleAddr2,
			Price:         sdk.NewInt(200),
			ReceivedAt:    now.Add(5 * time.Second),
		},
	}

	// check postPrice and rawPrices endpoints
	{
		// TX broadcast mode is "block" as using "sync" makes this test very unpredictable:
		//   it's not easy to find out when those TXs are Delivered
		prevBlockHeight := ct.WaitForNextBlocks(1)
		for _, postPrice := range postPrices {
			req, _ := ct.RestTxOraclePostPrice(postPrice.OracleName, postPrice.AssetCode, postPrice.Price, postPrice.ReceivedAt)
			req.CheckSucceeded()
		}
		curBlockHeight := ct.WaitForNextBlocks(1)

		// rawPrices could be stored in [prevBlockHeight : curBlockHeight], so we need to find them
		rawPrices := make([]oracle.PostedPrice, 0)
		for blockHeight := prevBlockHeight; blockHeight <= curBlockHeight; blockHeight++ {
			req, respMsg := ct.RestQueryOracleRawPrices(ct.DefAssetCode, blockHeight)
			req.CheckSucceeded()

			if len(*respMsg) > 0 {
				rawPrices = append(rawPrices, *respMsg...)
			}
		}

		require.Len(t, rawPrices, len(postPrices))
		for i, rawPrice := range rawPrices {
			postPrice := postPrices[i]
			require.Equal(t, rawPrice.AssetCode, postPrice.AssetCode)
			require.Equal(t, postPrice.OracleAddress, rawPrice.OracleAddress.String())
			require.True(t, rawPrice.Price.Equal(postPrice.Price))
			require.True(t, rawPrice.ReceivedAt.Equal(postPrice.ReceivedAt))
		}
	}

	// check rawPrices endpoint (invalid inputs)
	{
		// blockHeight without rawPrices
		{
			req, respMsg := ct.RestQueryOracleRawPrices(ct.DefAssetCode, 1)
			req.CheckSucceeded()

			require.Empty(t, *respMsg)
		}

		// non-existing assetCode
		{
			req, _ := ct.RestQueryOracleRawPrices("non_existing_asset", 1)
			req.CheckFailed(http.StatusNotFound, sdkErrors.ErrUnknownRequest)
		}
	}

	// check price endpoint
	{
		req, respMsg := ct.RestQueryOraclePrice(ct.DefAssetCode)
		req.CheckSucceeded()

		require.True(t, respMsg.Price.Equal(postPrices[1].Price))
		require.True(t, respMsg.ReceivedAt.Equal(postPrices[1].ReceivedAt))

		// check invalid inputs
		{
			// non-existing assetCode
			{
				req, _ := ct.RestQueryOraclePrice("non_existing_asset")
				req.CheckFailed(http.StatusNotFound, sdkErrors.ErrUnknownRequest)
			}
		}
	}
}

func Test_POARest(t *testing.T) {
	t.Parallel()
	ct := cliTester.New(t, false)
	defer ct.Close()
	ct.StartRestServer(false)

	// get all validators
	accs := make(map[string]cliTester.CLIAccount, 0)
	for _, acc := range ct.Accounts {
		if acc.IsPOAValidator {
			accs[acc.Address] = *acc
		}
	}

	// check getValidators endpoint
	{
		req, respMsg := ct.RestQueryPoaValidators()
		req.CheckSucceeded()

		require.Equal(t, len(accs), len(respMsg.Validators))
		for idx := range respMsg.Validators {
			sdkAddr := respMsg.Validators[idx].Address.String()
			ethAddr := respMsg.Validators[idx].EthAddress

			require.Contains(t, accs, sdkAddr)
			require.Equal(t, accs[sdkAddr].EthAddress, ethAddr)
		}
	}
}

func Test_VMRest(t *testing.T) {
	t.Parallel()
	ct := cliTester.New(t, false)
	defer ct.Close()
	ct.StartRestServer(false)

	vmGenState := vm.GenesisState{}
	// read default writeSets
	{
		file, err := os.Open(os.ExpandEnv(cliTester.DefVmWriteSetsPath))
		require.NoError(t, err, "open default writeSets file")

		jsonContent, err := ioutil.ReadAll(file)
		require.NoError(t, err, "reading default writeSets file")

		require.NoError(t, ct.Cdc.UnmarshalJSON(jsonContent, &vmGenState), "unmarshal default writeSets file")

		file.Close()
	}

	// check data endpoint
	{
		writeSet := vmGenState.WriteSet[0]
		req, respMsg := ct.RestQueryVMGetData(writeSet.Address, writeSet.Path)
		req.CheckSucceeded()

		require.Equal(t, writeSet.Value, respMsg.Value)

		// check invalid inputs
		{
			// invalid accAddress
			{
				req, _ := ct.RestQueryVMGetData("non-valid-addr", writeSet.Path)
				req.CheckFailed(http.StatusUnprocessableEntity, nil)
			}

			// invalid path
			{
				req, _ := ct.RestQueryVMGetData(writeSet.Address, "non-valid-path")
				req.CheckFailed(http.StatusUnprocessableEntity, nil)
			}
		}
	}
}

func Test_MarketsREST(t *testing.T) {
	t.Parallel()
	ct := cliTester.New(t, false)
	defer ct.Close()
	ct.StartRestServer(false)

	ownerName := "validator1"

	// add markets
	{
		r1, _ := ct.RestTxMarketsAdd(ownerName, cliTester.DenomBTC, cliTester.DenomDFI)
		r1.CheckSucceeded()
		r2, _ := ct.RestTxMarketsAdd(ownerName, cliTester.DenomETH, cliTester.DenomDFI)
		r2.CheckSucceeded()
	}

	// check addMarket Tx
	{
		// non-existing currency
		{
			r, _ := ct.RestTxMarketsAdd(ownerName, cliTester.DenomBTC, "atom")
			r.CheckFailed(http.StatusOK, marketTypes.ErrWrongAssetDenom)
		}

		// already existing market
		{
			r, _ := ct.RestTxMarketsAdd(ownerName, cliTester.DenomBTC, cliTester.DenomDFI)
			r.CheckFailed(http.StatusOK, marketTypes.ErrMarketExists)
		}
	}

	// check market query
	{
		// non-existing marketID
		{
			r, _ := ct.RestQueryMarket(dnTypes.NewIDFromUint64(10))
			r.CheckFailed(http.StatusInternalServerError, marketTypes.ErrWrongID)
		}

		// existing marketID (btc-dfi)
		{
			r, market := ct.RestQueryMarket(dnTypes.NewIDFromUint64(0))
			r.CheckSucceeded()

			require.Equal(t, market.ID.UInt64(), uint64(0))
			require.Equal(t, market.BaseAssetDenom, cliTester.DenomBTC)
			require.Equal(t, market.QuoteAssetDenom, cliTester.DenomDFI)
		}
	}

	// check list query
	{
		// all markets
		{
			r, markets := ct.RestQueryMarkets(-1, -1, nil, nil)
			r.CheckSucceeded()

			require.Len(t, *markets, 2)
			require.Equal(t, (*markets)[0].ID.UInt64(), uint64(0))
			require.Equal(t, (*markets)[0].BaseAssetDenom, cliTester.DenomBTC)
			require.Equal(t, (*markets)[1].ID.UInt64(), uint64(1))
			require.Equal(t, (*markets)[1].BaseAssetDenom, cliTester.DenomETH)
		}

		// check page / limit parameters
		{
			// page 1, limit 1
			rP1L1, marketsP1L1 := ct.RestQueryMarkets(1, 1, nil, nil)
			rP1L1.CheckSucceeded()

			require.Len(t, *marketsP1L1, 1)

			// page 2, limit 1
			rP2L1, marketsP2L1 := ct.RestQueryMarkets(1, 1, nil, nil)
			rP2L1.CheckSucceeded()

			require.Len(t, *marketsP2L1, 1)

			// page 2, limit 10 (no markets)
			rP2L10, marketsP2L10 := ct.RestQueryMarkets(2, 10, nil, nil)
			rP2L10.CheckSucceeded()

			require.Empty(t, *marketsP2L10)
		}

		// check baseDenom filter
		{
			baseDenom := cliTester.DenomETH
			r, markets := ct.RestQueryMarkets(-1, -1, &baseDenom, nil)
			r.CheckSucceeded()

			require.Len(t, *markets, 1)
			require.Equal(t, (*markets)[0].BaseAssetDenom, baseDenom)
		}

		// check quoteDenom filter
		{
			quoteDenom := cliTester.DenomDFI
			r, markets := ct.RestQueryMarkets(-1, -1, nil, &quoteDenom)
			r.CheckSucceeded()

			require.Len(t, *markets, 2)
			require.Equal(t, (*markets)[0].QuoteAssetDenom, quoteDenom)
			require.Equal(t, (*markets)[1].QuoteAssetDenom, quoteDenom)
		}

		// check multiple filters
		{
			baseDeno := cliTester.DenomBTC
			quoteDenom := cliTester.DenomDFI
			r, markets := ct.RestQueryMarkets(-1, -1, &baseDeno, &quoteDenom)
			r.CheckSucceeded()

			require.Len(t, *markets, 1)
		}
	}
}

func Test_OrdersREST(t *testing.T) {
	const (
		DecimalsDFI = "1000000000000000000"
		DecimalsETH = "1000000000000000000"
		DecimalsBTC = "100000000"
	)

	oneDfi := sdk.NewUintFromString(DecimalsDFI)
	oneBtc := sdk.NewUintFromString(DecimalsBTC)
	oneEth := sdk.NewUintFromString(DecimalsETH)
	accountBalances := []cliTester.StringPair{
		{
			Key:   cliTester.DenomBTC,
			Value: sdk.NewUint(10000).Mul(oneBtc).String(),
		},
		{
			Key:   cliTester.DenomETH,
			Value: sdk.NewUint(100000000).Mul(oneEth).String(),
		},
		{
			Key:   cliTester.DenomDFI,
			Value: sdk.NewUint(100000000).Mul(oneDfi).String(),
		},
	}
	accountOpts := []cliTester.AccountOption{
		{Name: "client1", Balances: accountBalances},
		{Name: "client2", Balances: accountBalances},
	}

	t.Parallel()
	ct := cliTester.New(
		t,
		false,
		cliTester.AccountsOption(accountOpts...),
	)
	defer ct.Close()
	ct.StartRestServer(false)

	ownerName1, ownerName2 := accountOpts[0].Name, accountOpts[1].Name
	ownerAddr1, ownerAddr2 := ct.Accounts[ownerName1].Address, ct.Accounts[ownerName2].Address
	marketID0, marketID1 := dnTypes.NewIDFromUint64(0), dnTypes.NewIDFromUint64(1)

	// add market
	{
		r1, _ := ct.RestTxMarketsAdd(ownerName1, cliTester.DenomBTC, cliTester.DenomDFI)
		r1.CheckSucceeded()
		r2, _ := ct.RestTxMarketsAdd(ownerName1, cliTester.DenomETH, cliTester.DenomDFI)
		r2.CheckSucceeded()
	}

	// check AddOrder Tx
	{
		// invalid marketID
		{
			r, _ := ct.RestTxOrdersPostOrder(ownerName1, dnTypes.NewIDFromUint64(2), orderTypes.AskDirection, sdk.OneUint(), sdk.OneUint(), 60)
			r.CheckFailed(http.StatusOK, orderTypes.ErrWrongMarketID)
		}
	}

	// add orders
	inputOrders := []struct {
		MarketID     dnTypes.ID
		OwnerName    string
		OwnerAddress string
		Direction    orderTypes.Direction
		Price        sdk.Uint
		Quantity     sdk.Uint
		TtlInSec     uint64
	}{
		{
			MarketID:     marketID0,
			OwnerName:    ownerName1,
			OwnerAddress: ownerAddr1,
			Direction:    orderTypes.BidDirection,
			Price:        sdk.NewUintFromString("10000000000000000000"),
			Quantity:     sdk.NewUintFromString("100000000"),
			TtlInSec:     60,
		},
		{
			MarketID:     marketID0,
			OwnerName:    ownerName2,
			OwnerAddress: ownerAddr2,
			Direction:    orderTypes.BidDirection,
			Price:        sdk.NewUintFromString("20000000000000000000"),
			Quantity:     sdk.NewUintFromString("200000000"),
			TtlInSec:     90,
		},
		{
			MarketID:     marketID0,
			OwnerName:    ownerName1,
			OwnerAddress: ownerAddr1,
			Direction:    orderTypes.AskDirection,
			Price:        sdk.NewUintFromString("50000000000000000000"),
			Quantity:     sdk.NewUintFromString("500000000"),
			TtlInSec:     60,
		},
		{
			MarketID:     marketID0,
			OwnerName:    ownerName2,
			OwnerAddress: ownerAddr2,
			Direction:    orderTypes.AskDirection,
			Price:        sdk.NewUintFromString("60000000000000000000"),
			Quantity:     sdk.NewUintFromString("600000000"),
			TtlInSec:     90,
		},
		{
			MarketID:     marketID1,
			OwnerName:    ownerName1,
			OwnerAddress: ownerAddr1,
			Direction:    orderTypes.AskDirection,
			Price:        sdk.NewUintFromString("10000000000000000000"),
			Quantity:     sdk.NewUintFromString("100000000"),
			TtlInSec:     30,
		},
	}
	for _, input := range inputOrders {
		r, _ := ct.RestTxOrdersPostOrder(input.OwnerName, input.MarketID, input.Direction, input.Price, input.Quantity, input.TtlInSec)
		r.CheckSucceeded()
	}

	// check orders added
	{
		for i, input := range inputOrders {
			orderID := dnTypes.NewIDFromUint64(uint64(i))
			q, order := ct.RestQueryOrder(orderID)
			q.CheckSucceeded()

			require.True(t, order.ID.Equal(orderID), "order %d: ID", i)
			require.True(t, order.Market.ID.Equal(input.MarketID), "order %d: MarketID", i)
			require.Equal(t, order.Owner.String(), input.OwnerAddress, "order %d: Owner", i)
			require.True(t, order.Direction.Equal(input.Direction), "order %d: Direction", i)
			require.True(t, order.Price.Equal(input.Price), "order %d: Price", i)
			require.True(t, order.Quantity.Equal(input.Quantity), "order %d: Quantity", i)
			require.Equal(t, order.Ttl, time.Duration(input.TtlInSec)*time.Second, "order %d: Ttl", i)
		}
	}

	// check list query
	{
		// request all
		{
			q, orders := ct.RestQueryOrders(-1, -1, nil, nil, nil)
			q.CheckSucceeded()

			require.Len(t, *orders, len(inputOrders))
		}

		// check page / limit parameters
		{
			// page 1, limit 1
			qP1L1, ordersP1L1 := ct.RestQueryOrders(1, 1, nil, nil, nil)
			qP1L1.CheckSucceeded()

			require.Len(t, *ordersP1L1, 1)

			// page 2, limit 1
			qP2L1, ordersP2L1 := ct.RestQueryOrders(1, 1, nil, nil, nil)
			qP2L1.CheckSucceeded()

			require.Len(t, *ordersP2L1, 1)

			// page 2, limit 10 (no orders)
			qP2L10, ordersP2L10 := ct.RestQueryOrders(2, 10, nil, nil, nil)
			qP2L10.CheckSucceeded()

			require.Empty(t, *ordersP2L10)
		}

		// check marketID filter
		{
			market0Count, market1Count := 0, 0
			for _, input := range inputOrders {
				if input.MarketID.UInt64() == 0 {
					market0Count++
				}
				if input.MarketID.UInt64() == 1 {
					market1Count++
				}
			}

			q0, orders0 := ct.RestQueryOrders(-1, -1, &marketID0, nil, nil)
			q0.CheckSucceeded()

			require.Len(t, *orders0, market0Count)

			q1, orders1 := ct.RestQueryOrders(-1, -1, &marketID1, nil, nil)
			q1.CheckSucceeded()

			require.Len(t, *orders1, market1Count)
		}

		// check direction filter
		{
			askCount, bidCount := 0, 0
			for _, input := range inputOrders {
				if input.Direction.Equal(orderTypes.AskDirection) {
					askCount++
				}
				if input.Direction.Equal(orderTypes.BidDirection) {
					bidCount++
				}
			}

			askDirection := orderTypes.AskDirection
			qAsk, ordersAsk := ct.RestQueryOrders(-1, -1, nil, &askDirection, nil)
			qAsk.CheckSucceeded()

			require.Len(t, *ordersAsk, askCount)

			bidDirection := orderTypes.BidDirection
			qBid, ordersBid := ct.RestQueryOrders(-1, -1, nil, &bidDirection, nil)
			qBid.CheckSucceeded()

			require.Len(t, *ordersBid, bidCount)
		}

		// check owner filter
		{
			client1Count, client2Count := 0, 0
			for _, input := range inputOrders {
				if input.OwnerAddress == ownerAddr1 {
					client1Count++
				}
				if input.OwnerAddress == ownerAddr2 {
					client2Count++
				}
			}

			q1, orders1 := ct.RestQueryOrders(-1, -1, nil, nil, &ownerAddr1)
			q1.CheckSucceeded()

			require.Len(t, *orders1, client1Count)

			q2, orders2 := ct.RestQueryOrders(-1, -1, nil, nil, &ownerAddr2)
			q2.CheckSucceeded()

			require.Len(t, *orders2, client2Count)
		}

		// check multiple filters
		{
			marketID := marketID0
			owner := ownerAddr1
			direction := orderTypes.AskDirection
			count := 0
			for _, input := range inputOrders {
				if input.MarketID.Equal(marketID) && input.OwnerAddress == owner && input.Direction == direction {
					count++
				}
			}

			q, orders := ct.RestQueryOrders(-1, -1, &marketID, &direction, &owner)
			q.CheckSucceeded()

			require.Len(t, *orders, count)
		}
	}

	// revoke order
	{
		orderIdx := len(inputOrders) - 1
		orderID := dnTypes.NewIDFromUint64(uint64(orderIdx))
		inputOrder := inputOrders[orderIdx]
		r, _ := ct.RestTxOrdersRevokeOrder(inputOrder.OwnerName, orderID)
		r.CheckSucceeded()

		q, _ := ct.RestQueryOrder(orderID)
		q.CheckFailed(http.StatusInternalServerError, orderTypes.ErrWrongOrderID)
		inputOrders = inputOrders[:len(inputOrders)-2]
	}

	// check RevokeOrder Tx
	{
		// non-existing orderID
		{
			r, _ := ct.RestTxOrdersRevokeOrder(ownerName1, dnTypes.NewIDFromUint64(10))
			r.CheckFailed(http.StatusOK, orderTypes.ErrWrongOrderID)
		}

		// wrong owner (not an order owner)
		{
			r, _ := ct.RestTxOrdersRevokeOrder("validator1", dnTypes.NewIDFromUint64(0))
			r.CheckFailed(http.StatusOK, orderTypes.ErrWrongOwner)
		}
	}
}
