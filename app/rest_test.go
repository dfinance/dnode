// +build rest

package app

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	cliTester "github.com/dfinance/dnode/helpers/tests/clitester"
	ccTypes "github.com/dfinance/dnode/x/currencies/types"
	msMsgs "github.com/dfinance/dnode/x/multisig/msgs"
	msTypes "github.com/dfinance/dnode/x/multisig/types"
	"github.com/dfinance/dnode/x/oracle"
	poaMsgs "github.com/dfinance/dnode/x/poa/msgs"
	poaTypes "github.com/dfinance/dnode/x/poa/types"
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
	ct.WaitForNextBlocks(1)
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
				req.CheckFailed(http.StatusInternalServerError, ccTypes.ErrWrongIssueID(""))
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
				req.CheckFailed(http.StatusInternalServerError, ccTypes.ErrNotExistCurrency(""))
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
	ct.WaitForNextBlocks(1)
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
	ct.WaitForNextBlocks(1)
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
	r := NewRestTester(t, false)
	defer r.Close()

	senderAddr, senderPrivKey := r.Accounts[0].Address, r.PrivKeys[0]
	calls := msTypes.CallsResp{}
	msgIDs := make([]string, 0)

	// submit remove validator call (1st one)
	{
		senderAcc := GetAccountCheckTx(r.App, senderAddr)
		targetValidator := r.Accounts[len(r.Accounts)-1]

		removeMsg := poaMsgs.NewMsgRemoveValidator(targetValidator.Address, senderAcc.GetAddress())
		msgID := fmt.Sprintf("removeValidator:%s", targetValidator.Address)
		msgIDs = append(msgIDs, msgID)

		submitMsg := msMsgs.NewMsgSubmitCall(removeMsg, msgID, senderAcc.GetAddress())
		tx := genTx([]sdk.Msg{submitMsg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
		CheckDeliverTx(t, r.App, tx)

		CheckRunQuery(t, r.App, nil, queryMsGetCallsPath, &calls)
		require.Equal(t, 1, len(calls))
		require.Equal(t, 1, len(calls[0].Votes))
	}

	// submit remove validator call (2nd one)
	{
		senderAcc := GetAccountCheckTx(r.App, senderAddr)
		targetValidator := r.Accounts[len(r.Accounts)-2]

		removeMsg := poaMsgs.NewMsgRemoveValidator(targetValidator.Address, senderAcc.GetAddress())
		msgID := fmt.Sprintf("removeValidator:%s", targetValidator)
		msgIDs = append(msgIDs, msgID)

		submitMsg := msMsgs.NewMsgSubmitCall(removeMsg, msgID, senderAcc.GetAddress())
		tx := genTx([]sdk.Msg{submitMsg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
		CheckDeliverTx(t, r.App, tx)

		CheckRunQuery(t, r.App, nil, queryMsGetCallsPath, &calls)
		require.Equal(t, 2, len(calls))
		require.Equal(t, 1, len(calls[1].Votes))
	}

	// check getCalls endpoint
	{
		reqSubPath := fmt.Sprintf("%s/calls", msTypes.ModuleName)
		respMsg := msTypes.CallsResp{}

		r.Request("GET", reqSubPath, nil, nil, &respMsg, true)
		require.Len(t, respMsg, 2)
		for i, call := range respMsg {
			require.Len(t, call.Votes, 1)
			require.Equal(t, senderAddr, call.Votes[0])
			require.Equal(t, uint64(i), call.Call.MsgID)
			require.Equal(t, senderAddr, call.Call.Creator)
			require.Equal(t, msgIDs[i], call.Call.UniqueID)
		}
	}

	// check getCall endpoint
	{
		reqSubPath := fmt.Sprintf("%s/call/%d", msTypes.ModuleName, 0)
		respMsg := msTypes.CallResp{}

		r.Request("GET", reqSubPath, nil, nil, &respMsg, true)
		require.Len(t, respMsg.Votes, 1)
		require.Equal(t, senderAddr, respMsg.Votes[0])
		require.Equal(t, uint64(0), respMsg.Call.MsgID)
		require.Equal(t, senderAddr, respMsg.Call.Creator)
		require.Equal(t, msgIDs[0], respMsg.Call.UniqueID)
	}

	// check getCall endpoint (invalid "id")
	{
		reqSubPath := fmt.Sprintf("%s/call/-1", msTypes.ModuleName)

		respCode, _ := r.Request("GET", reqSubPath, nil, nil, nil, false)
		r.CheckError(http.StatusInternalServerError, respCode, nil, nil)
	}

	// check getCall endpoint (non-existing "id")
	{
		reqSubPath := fmt.Sprintf("%s/call/2", msTypes.ModuleName)

		respCode, respBytes := r.Request("GET", reqSubPath, nil, nil, nil, false)
		r.CheckError(http.StatusInternalServerError, respCode, msTypes.ErrWrongCallId(0), respBytes)
	}

	// check getCallByUnique endpoint
	{
		reqSubPath := fmt.Sprintf("%s/unique/%s", msTypes.ModuleName, msgIDs[0])
		respMsg := msTypes.CallResp{}

		r.Request("GET", reqSubPath, nil, nil, &respMsg, true)
		require.Len(t, respMsg.Votes, 1)
		require.Equal(t, senderAddr, respMsg.Votes[0])
		require.Equal(t, uint64(0), respMsg.Call.MsgID)
		require.Equal(t, senderAddr, respMsg.Call.Creator)
		require.Equal(t, msgIDs[0], respMsg.Call.UniqueID)
	}

	// check getCallByUnique endpoint (non-existing "unique")
	{
		reqSubPath := fmt.Sprintf("%s/unique/non-existing-UNIQUE", msTypes.ModuleName)

		respCode, respBytes := r.Request("GET", reqSubPath, nil, nil, nil, false)
		r.CheckError(http.StatusInternalServerError, respCode, msTypes.ErrNotFoundUniqueID(""), respBytes)
	}
}

func Test_OracleRest(t *testing.T) {
	r := NewRestTester(t, false)
	defer r.Close()

	// check getAssets endpoint
	{
		reqSubPath := fmt.Sprintf("%s/assets", oracle.ModuleName)
		respMsg := oracle.Assets{}

		r.Request("GET", reqSubPath, nil, nil, &respMsg, true)
		require.Len(t, respMsg, 1)
		require.Equal(t, r.DefaultAssetCode, respMsg[0].AssetCode)
		require.Len(t, respMsg[0].Oracles, 2)
		require.True(t, r.Accounts[0].Address.Equals(respMsg[0].Oracles[0].Address))
		require.True(t, r.Accounts[1].Address.Equals(respMsg[0].Oracles[1].Address))
		require.True(t, respMsg[0].Active)
	}

	now := time.Now()
	postPrices := []struct {
		AssetCode     string
		SenderIdx     uint
		OracleAddress sdk.AccAddress
		Price         sdk.Int
		ReceivedAt    time.Time
	}{
		{
			AssetCode:     r.DefaultAssetCode,
			SenderIdx:     0,
			OracleAddress: r.Accounts[0].Address,
			Price:         sdk.NewInt(100),
			ReceivedAt:    now,
		},
		{
			AssetCode:     r.DefaultAssetCode,
			SenderIdx:     1,
			OracleAddress: r.Accounts[1].Address,
			Price:         sdk.NewInt(200),
			ReceivedAt:    now.Add(5 * time.Second),
		},
	}

	// check postPrice and rawPrices endpoints
	{
		prevBlockHeight := r.WaitForNextBlock()
		for _, postPrice := range postPrices {
			reqMsg := oracle.NewMsgPostPrice(postPrice.OracleAddress, postPrice.AssetCode, postPrice.Price, postPrice.ReceivedAt)

			r.TxSyncRequest(postPrice.SenderIdx, reqMsg, true)
		}
		curBlockHeight := r.WaitForNextBlock()

		// rawPrices could be stored in [prevBlockHeight : curBlockHeight], so we need to find them
		rawPrices := make([]oracle.PostedPrice, 0)
		for blockHeight := prevBlockHeight; blockHeight <= curBlockHeight; blockHeight++ {
			reqSubPath := fmt.Sprintf("%s/rawprices/%s/%d", oracle.ModuleName, r.DefaultAssetCode, blockHeight)

			r.Request("GET", reqSubPath, nil, nil, &rawPrices, true)
			if len(rawPrices) > 0 {
				return
			}
		}

		require.Len(t, rawPrices, len(postPrices))
		for i, rawPrice := range rawPrices {
			postPrice := postPrices[i]
			require.Equal(t, rawPrice.AssetCode, postPrice.AssetCode)
			require.True(t, rawPrice.OracleAddress.Equals(postPrice.OracleAddress))
			require.True(t, rawPrice.Price.Equal(postPrice.Price))
			require.True(t, rawPrice.ReceivedAt.Equal(postPrice.ReceivedAt))
		}
	}

	// check rawPrices endpoint (invalid arguments)
	{
		// blockHeight
		{
			reqSubPath := fmt.Sprintf("%s/rawprices/%s/%d", oracle.ModuleName, r.DefaultAssetCode, 1)
			rawPrices := make([]oracle.PostedPrice, 0)

			r.Request("GET", reqSubPath, nil, nil, &rawPrices, true)
			require.Empty(t, rawPrices)
		}
		// assetCode
		{
			reqSubPath := fmt.Sprintf("%s/rawprices/%s/%d", oracle.ModuleName, "non_existing_asset", 1)

			rcvCode, rcvBytes := r.Request("GET", reqSubPath, nil, nil, nil, false)
			r.CheckError(http.StatusNotFound, rcvCode, sdk.ErrUnknownRequest(""), rcvBytes)
		}
	}

	// check price endpoint
	{
		reqSubPath := fmt.Sprintf("%s/currentprice/%s", oracle.ModuleName, r.DefaultAssetCode)
		avgPrice := postPrices[0].Price.Add(postPrices[1].Price).Quo(sdk.NewInt(2))
		price := oracle.CurrentPrice{}

		r.Request("GET", reqSubPath, nil, nil, &price, true)
		require.True(t, price.Price.Equal(avgPrice))
		require.False(t, price.ReceivedAt.Equal(postPrices[0].ReceivedAt))
		require.False(t, price.ReceivedAt.Equal(postPrices[1].ReceivedAt))
	}

	// check price endpoint (invalid arguments)
	{
		// assetCode
		{
			reqSubPath := fmt.Sprintf("%s/currentprice/%s", oracle.ModuleName, "non_existing_asset")

			rcvCode, rcvBytes := r.Request("GET", reqSubPath, nil, nil, nil, false)
			r.CheckError(http.StatusNotFound, rcvCode, sdk.ErrUnknownRequest(""), rcvBytes)
		}
	}
}

func Test_POARest(t *testing.T) {
	r := NewRestTester(t, false)
	defer r.Close()

	// check getValidators endpoint
	{
		reqSubPath := fmt.Sprintf("%s/validators", poaTypes.ModuleName)
		respMsg := poaTypes.ValidatorsConfirmations{}

		r.Request("GET", reqSubPath, nil, nil, &respMsg, true)
		require.Equal(t, len(r.Accounts), len(respMsg.Validators))
		for idx := range respMsg.Validators {
			require.Equal(t, r.Accounts[idx].GetAddress(), respMsg.Validators[idx].Address)
			require.Equal(t, "0x17f7D1087971dF1a0E6b8Dae7428E97484E32615", respMsg.Validators[idx].EthAddress)
		}
	}
}
