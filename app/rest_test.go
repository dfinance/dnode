// +build rest

package app

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"

	cliTester "github.com/dfinance/dnode/helpers/tests/clitester"
	ccTypes "github.com/dfinance/dnode/x/currencies/types"
	msTypes "github.com/dfinance/dnode/x/multisig/types"
	"github.com/dfinance/dnode/x/oracle"
)

func Test_CurrencyRest(t *testing.T) {
	ct := cliTester.New(t, false)
	defer ct.Close()
	ct.StartRestServer(false)

	recipientAddr := ct.Accounts["validator1"].Address
	curAmount, curDecimals, denom, issueId := sdk.NewInt(100), int8(0), "tstdnm", "issue1"
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
		require.Equal(t, ct.ChainID, respMsg.ChainID)
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
			require.Equal(t, ct.ChainID, destroy.ChainID)
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
