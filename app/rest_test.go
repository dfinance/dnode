// +build rest

package app

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	ccTypes "github.com/dfinance/dnode/x/currencies/types"
	msMsgs "github.com/dfinance/dnode/x/multisig/msgs"
	msTypes "github.com/dfinance/dnode/x/multisig/types"
	"github.com/dfinance/dnode/x/oracle"
	poaMsgs "github.com/dfinance/dnode/x/poa/msgs"
	poaTypes "github.com/dfinance/dnode/x/poa/types"
)

func Test_CurrencyRest(t *testing.T) {
	r := NewRestTester(t)
	defer r.Close()

	recipientIdx, recipientAddr, recipientPrivKey := uint(0), r.Accounts[0].Address, r.PrivKeys[0]
	curAmount, curDecimals, denom := sdk.NewInt(100), int8(0), currency1Symbol
	destroyAmounts := make([]sdk.Int, 0)

	// issue currency
	msgId, issueId := "1", "issue1"
	issueCurrency(t, r.App, denom, curAmount, curDecimals, msgId, issueId, recipientIdx, r.Accounts, r.PrivKeys, true)
	checkIssueExists(t, r.App, issueId, denom, curAmount, recipientAddr)
	checkCurrencyExists(t, r.App, denom, curAmount, curDecimals)

	// check getIssue endpoint
	{
		reqSubPath := fmt.Sprintf("%s/issue/%s", ccTypes.ModuleName, issueId)
		respMsg := &ccTypes.Issue{}

		r.Request("GET", reqSubPath, nil, nil, respMsg, true)
		require.Equal(t, denom, respMsg.Symbol)
		require.True(t, respMsg.Amount.Equal(curAmount))
		require.Equal(t, recipientAddr, respMsg.Recipient)
	}

	// check getIssue endpoint (invalid issueID)
	{
		reqSubPath := fmt.Sprintf("%s/issue/non_existing_ID", ccTypes.ModuleName)

		respCode, respBytes := r.Request("GET", reqSubPath, nil, nil, nil, false)
		r.CheckError(http.StatusInternalServerError, respCode, ccTypes.ErrWrongIssueID(""), respBytes)
	}

	// check getCurrency endpoint
	{
		reqSubPath := fmt.Sprintf("%s/currency/%s", ccTypes.ModuleName, denom)
		respMsg := &ccTypes.Currency{}

		r.Request("GET", reqSubPath, nil, nil, respMsg, true)
		require.Equal(t, denom, respMsg.Symbol)
		require.True(t, respMsg.Supply.Equal(curAmount))
		require.Equal(t, curDecimals, respMsg.Decimals)
	}

	// check getCurrency endpoint (invalid symbol)
	{
		reqSubPath := fmt.Sprintf("%s/currency/non_existing_symbol", ccTypes.ModuleName)

		respCode, respBytes := r.Request("GET", reqSubPath, nil, nil, nil, false)
		r.CheckError(http.StatusInternalServerError, respCode, ccTypes.ErrNotExistCurrency(""), respBytes)
	}

	// check getDestroys endpoint (no destroys)
	{
		reqSubPath := fmt.Sprintf("%s/destroys/1", ccTypes.ModuleName)
		respMsg := ccTypes.Destroys{}

		r.Request("GET", reqSubPath, nil, nil, &respMsg, true)
		require.Len(t, respMsg, 0)
	}

	// destroy currency
	newAmount := sdk.NewInt(50)
	curAmount = curAmount.Sub(newAmount)
	destroyCurrency(t, r.App, r.ChainId, denom, newAmount, recipientAddr, recipientPrivKey, true)
	checkCurrencyExists(t, r.App, denom, curAmount, 0)
	destroyAmounts = append(destroyAmounts, newAmount)

	// check getDestroy endpoint
	{
		reqSubPath := fmt.Sprintf("%s/destroy/%d", ccTypes.ModuleName, 0)
		respMsg := &ccTypes.Destroy{}

		r.Request("GET", reqSubPath, nil, nil, respMsg, true)
		require.Equal(t, int64(0), respMsg.ID.Int64())
		require.Equal(t, r.ChainId, respMsg.ChainID)
		require.Equal(t, denom, respMsg.Symbol)
		require.True(t, respMsg.Amount.Equal(newAmount))
		require.Equal(t, recipientAddr, respMsg.Spender)
		require.Equal(t, recipientAddr.String(), respMsg.Recipient)
	}

	// check getDestroy endpoint (invalid destroyID)
	{
		reqSubPath := fmt.Sprintf("%s/destroy/abc", ccTypes.ModuleName)

		respCode, _ := r.Request("GET", reqSubPath, nil, nil, nil, false)
		r.CheckError(http.StatusInternalServerError, respCode, nil, nil)
	}

	// check getDestroy endpoint (non-existing destroyID)
	{
		reqSubPath := fmt.Sprintf("%s/destroy/1", ccTypes.ModuleName)
		respMsg := &ccTypes.Destroy{}

		r.Request("GET", reqSubPath, nil, nil, respMsg, true)
		require.Empty(t, respMsg.ChainID)
		require.Empty(t, respMsg.Symbol)
		require.True(t, respMsg.Amount.IsZero())
	}

	// destroy currency once more
	newAmount = sdk.NewInt(25)
	curAmount = curAmount.Sub(newAmount)
	destroyCurrency(t, r.App, r.ChainId, denom, newAmount, recipientAddr, recipientPrivKey, true)
	checkCurrencyExists(t, r.App, denom, curAmount, 0)
	destroyAmounts = append(destroyAmounts, newAmount)

	// check getDestroys endpoint
	{
		reqSubPath := fmt.Sprintf("%s/destroys/1", ccTypes.ModuleName)
		respMsg := ccTypes.Destroys{}

		r.Request("GET", reqSubPath, nil, nil, &respMsg, true)
		require.Len(t, respMsg, len(destroyAmounts))
		for i, amount := range destroyAmounts {
			destroy := respMsg[i]
			require.Equal(t, int64(i), destroy.ID.Int64())
			require.Equal(t, r.ChainId, destroy.ChainID)
			require.Equal(t, denom, destroy.Symbol)
			require.True(t, destroy.Amount.Equal(amount))
			require.Equal(t, recipientAddr, destroy.Spender)
			require.Equal(t, recipientAddr.String(), destroy.Recipient)
		}
	}

	// check getDestroys endpoint (invalid query values)
	{
		// invalid "page" value
		reqSubPath := fmt.Sprintf("%s/destroys/abc", ccTypes.ModuleName)

		respCode, _ := r.Request("GET", reqSubPath, nil, nil, nil, false)
		r.CheckError(http.StatusInternalServerError, respCode, nil, nil)

		// invalid "limit" value
		reqSubPath = fmt.Sprintf("%s/destroys/1", ccTypes.ModuleName)
		reqValues := url.Values{}
		reqValues.Set("limit", "abc")

		respCode, _ = r.Request("GET", reqSubPath, reqValues, nil, nil, false)
		r.CheckError(http.StatusInternalServerError, respCode, nil, nil)
	}
}

func Test_MSRest(t *testing.T) {
	r := NewRestTester(t)
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
	r := NewRestTester(t)
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
	r := NewRestTester(t)
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
