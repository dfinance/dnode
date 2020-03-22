package app

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"

	"github.com/dfinance/dnode/cmd/config"
	"github.com/dfinance/dnode/x/currencies"
	ccMsgs "github.com/dfinance/dnode/x/currencies/msgs"
	ccTypes "github.com/dfinance/dnode/x/currencies/types"
	msTypes "github.com/dfinance/dnode/x/multisig/types"
)

const (
	queryCurrencyGetIssuePath    = "/custom/currencies/" + currencies.QueryGetIssue
	queryCurrencyGetCurrencyPath = "/custom/currencies/" + currencies.QueryGetCurrency
	queryCurrencyGetDestroyPath  = "/custom/currencies/" + currencies.QueryGetDestroy
	queryCurrencyGetDestroysPath = "/custom/currencies/" + currencies.QueryGetDestroys
)

func Test_CurrencyHandlerIsMultisigOnly(t *testing.T) {
	app, server := newTestDnApp()
	defer app.CloseConnections()
	defer server.Stop()

	genCoins, err := sdk.ParseCoins("1000000000000000" + config.MainDenom)
	require.NoError(t, err)
	genValidators, _, _, genPrivKeys := CreateGenAccounts(7, genCoins)

	_, err = setGenesis(t, app, genValidators)
	require.NoError(t, err)

	// check module supports only multisig calls (using MSRouter)
	{
		senderAcc, senderPrivKey := GetAccountCheckTx(app, genValidators[0].Address), genPrivKeys[0]
		issueMsg := ccMsgs.NewMsgIssueCurrency(currency1Symbol, amount, 0, senderAcc.GetAddress(), issue1ID)
		tx := genTx([]sdk.Msg{issueMsg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
		CheckDeliverSpecificErrorTx(t, app, tx, sdk.ErrUnauthorized(""))
	}
}

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

func Test_CurrencyQueries(t *testing.T) {
	app, server := newTestDnApp()
	defer app.CloseConnections()
	defer server.Stop()

	genCoins, err := sdk.ParseCoins("1000000000000000" + config.MainDenom)
	require.NoError(t, err)
	genAccs, _, _, genPrivKeys := CreateGenAccounts(10, genCoins)

	_, err = setGenesis(t, app, genAccs)
	require.NoError(t, err)

	recipientIdx, recipientAddr, recipientPrivKey := uint(0), genAccs[0].Address, genPrivKeys[0]

	checkCurrencyQueryObj := func(obj ccTypes.Currency, symbol string, amount sdk.Int, decimals int8) {
		require.Equal(t, obj.Symbol, symbol)
		require.True(t, obj.Supply.Equal(amount))
		require.Equal(t, obj.Decimals, decimals)
	}

	checkDestroyQueryObj := func(obj ccTypes.Destroy, id int64, symbol string, amount sdk.Int, recipientAddr sdk.AccAddress) {
		require.Equal(t, int64(0), obj.ID.Int64())
		require.Equal(t, chainID, obj.ChainID)
		require.Equal(t, symbol, obj.Symbol)
		require.True(t, obj.Amount.Equal(amount))
		require.Equal(t, recipientAddr, obj.Spender)
	}

	// issue multiple currencies
	issueCurrency(t, app, currency1Symbol, amount, 0, "msg1", issue1ID, recipientIdx, genAccs, genPrivKeys, true)
	issueCurrency(t, app, currency2Symbol, amount, 0, "msg2", issue2ID, recipientIdx, genAccs, genPrivKeys, true)
	issueCurrency(t, app, currency3Symbol, amount, 0, "msg3", issue3ID, recipientIdx, genAccs, genPrivKeys, true)

	// check getCurrency query
	{
		checkSymbol := func(symbol string) {
			currencyObj := ccTypes.Currency{}
			CheckRunQuery(t, app, ccTypes.CurrencyReq{Symbol: symbol}, queryCurrencyGetCurrencyPath, &currencyObj)
			checkCurrencyQueryObj(currencyObj, symbol, amount, 0)
		}

		checkSymbol(currency1Symbol)
		checkSymbol(currency2Symbol)
		checkSymbol(currency3Symbol)
	}

	// destroy currency
	destroyCurrency(t, app, chainID, currency3Symbol, amount, recipientAddr, recipientPrivKey, true)
	destroyID := int64(0)

	// check destroys query with pagination
	{
		destroys := ccTypes.Destroys{}
		CheckRunQuery(t, app, ccTypes.DestroysReq{Page: sdk.NewInt(1), Limit: sdk.NewInt(1)}, queryCurrencyGetDestroysPath, &destroys)

		require.Len(t, destroys, 1)
		checkDestroyQueryObj(destroys[0], destroyID, currency3Symbol, amount, recipientAddr)
	}

	// check single destroy query
	{
		destroy := ccTypes.Destroy{}
		CheckRunQuery(t, app, ccTypes.DestroyReq{DestroyId: sdk.NewInt(destroyID)}, queryCurrencyGetDestroyPath, &destroy)

		checkDestroyQueryObj(destroy, destroyID, currency3Symbol, amount, recipientAddr)
	}

	// check non-existing currency query
	{
		CheckRunQuerySpecificError(t, app, ccTypes.CurrencyReq{Symbol: "non-existing-symbol"}, queryCurrencyGetCurrencyPath, ccTypes.ErrNotExistCurrency(""))
	}
}

func Test_CurrencyIssue(t *testing.T) {
	app, server := newTestDnApp()
	defer app.CloseConnections()
	defer server.Stop()

	genCoins, err := sdk.ParseCoins("1000000000000000" + config.MainDenom)
	require.NoError(t, err)
	genAccs, _, _, genPrivKeys := CreateGenAccounts(10, genCoins)

	_, err = setGenesis(t, app, genAccs)
	require.NoError(t, err)

	recipientIdx, recipientAddr := uint(0), genAccs[0].Address
	curAmount, curDecimals, denom := amount, int8(0), currency1Symbol

	// check currency is issued
	{
		msgId, issueId := "1", "issue1"

		issueCurrency(t, app, denom, curAmount, curDecimals, msgId, issueId, recipientIdx, genAccs, genPrivKeys, true)
		checkIssueExists(t, app, issueId, denom, curAmount, recipientAddr)
		checkCurrencyExists(t, app, denom, curAmount, curDecimals)
		checkRecipientCoins(t, app, recipientAddr, denom, curAmount, curDecimals)
	}

	// check currency supply increased
	{
		msgId, issueId := "2", "issue2"
		newAmount := sdk.NewInt(200)
		curAmount = curAmount.Add(newAmount)

		issueCurrency(t, app, denom, newAmount, curDecimals, msgId, issueId, recipientIdx, genAccs, genPrivKeys, true)
		checkIssueExists(t, app, issueId, denom, newAmount, recipientAddr)
		checkCurrencyExists(t, app, denom, curAmount, curDecimals)
		checkRecipientCoins(t, app, recipientAddr, denom, curAmount, curDecimals)
	}

	// check currency issue for existing currency with different decimals
	{
		msgId, issueId := "3", "issue3"

		res := issueCurrency(t, app, denom, sdk.OneInt(), curDecimals+1, msgId, issueId, recipientIdx, genAccs, genPrivKeys, false)
		CheckResultError(t, ccTypes.ErrIncorrectDecimals(0, 0, ""), res)
	}

	// check zero amount currency issue
	{
		msgId, issueId, denom := "non-existing-msgID", "non-existing-issue", "non-existing-denom"

		curAmount, curDecimals := sdk.ZeroInt(), int8(0)
		res := issueCurrency(t, app, denom, curAmount, curDecimals, msgId, issueId, recipientIdx, genAccs, genPrivKeys, false)
		CheckResultError(t, ccTypes.ErrWrongAmount(""), res)
	}

	// check amount with negative decimals currency issue
	{
		msgId, issueId, denom := "non-existing-msgID", "non-existing-issue", "non-existing-denom"

		curAmount, curDecimals := sdk.OneInt(), int8(-1)
		res := issueCurrency(t, app, denom, curAmount, curDecimals, msgId, issueId, recipientIdx, genAccs, genPrivKeys, false)
		CheckResultError(t, ccTypes.ErrWrongDecimals(0), res)
	}

	// check currency issue with wrong symbol
	{
		msgId, issueId := "non-existing-msgID", "non-existing-issue"

		res := issueCurrency(t, app, "", amount, 0, msgId, issueId, recipientIdx, genAccs, genPrivKeys, false)
		CheckResultError(t, ccTypes.ErrWrongSymbol(""), res)
	}

	// check currency issue with the same issueID
	{
		msgId, issueId := "non-existing-msgID", "issue1"

		res := issueCurrency(t, app, denom, amount, 0, msgId, issueId, recipientIdx, genAccs, genPrivKeys, false)
		CheckResultError(t, ccTypes.ErrExistsIssue(""), res)
	}

	// check currency issue with already existing uniqueMsgID
	{
		msgId, issueId := "1", "non-existing-issue"

		res := issueCurrency(t, app, denom, amount, 0, msgId, issueId, recipientIdx, genAccs, genPrivKeys, false)
		CheckResultError(t, msTypes.ErrNotUniqueID(""), res)
	}

	// check currency issue with negative amount
	{
		msgId, issueId := "non-existing-msgID", "non-existing-issue"

		res := issueCurrency(t, app, denom, sdk.NewInt(-1), 0, msgId, issueId, recipientIdx, genAccs, genPrivKeys, false)
		require.Equal(t, sdk.CodespaceRoot, res.Codespace)
		require.Equal(t, sdk.CodeInternal, res.Code)
		require.Contains(t, res.Log, "negative coin amount")
	}

	// check currency issue with invalid denom
	{
		msgId, issueId := "non-existing-msgID", "non-existing-issue"
		invalidDenom := "1"

		res := issueCurrency(t, app, invalidDenom, sdk.NewInt(1), 0, msgId, issueId, recipientIdx, genAccs, genPrivKeys, false)
		require.Equal(t, sdk.CodespaceRoot, res.Codespace)
		require.Equal(t, sdk.CodeInternal, res.Code)
		require.Contains(t, res.Log, "invalid denom")
	}
}

func Test_CurrencyIssueHugeAmount(t *testing.T) {
	app, server := newTestDnApp()
	defer app.CloseConnections()
	defer server.Stop()

	genCoins, err := sdk.ParseCoins("1000000000000000" + config.MainDenom)
	require.NoError(t, err)
	genAccs, _, _, genPrivKeys := CreateGenAccounts(10, genCoins)

	_, err = setGenesis(t, app, genAccs)
	require.NoError(t, err)

	recipientIdx, recipientAddr := uint(0), genAccs[0].Address

	// check huge amount currency issue (max value for u128)
	{
		msgId, issueId, denom := "1", "issue1", currency1Symbol

		hugeAmount, ok := sdk.NewIntFromString("100000000000000000000000000000000000000")
		require.True(t, ok, "hugeAmount creation ()")

		issueCurrency(t, app, denom, hugeAmount, 0, msgId, issueId, recipientIdx, genAccs, genPrivKeys, true)
		checkIssueExists(t, app, issueId, denom, hugeAmount, recipientAddr)
		checkCurrencyExists(t, app, denom, hugeAmount, 0)
		checkRecipientCoins(t, app, recipientAddr, denom, hugeAmount, 0)
	}

	// check huge amount currency issue (that worked before u128)
	{
		msgId, issueId, denom := "2", "issue2", currency2Symbol

		hugeAmount, ok := sdk.NewIntFromString("1000000000000000000000000000000000000000000000")
		require.True(t, ok, "hugeAmount creation ()")

		issueCurrency(t, app, denom, hugeAmount, 0, msgId, issueId, recipientIdx, genAccs, genPrivKeys, true)
		checkIssueExists(t, app, issueId, denom, hugeAmount, recipientAddr)
		checkCurrencyExists(t, app, denom, hugeAmount, 0)

		require.Panics(t, func() {
			app.bankKeeper.GetCoins(GetContext(app, true), recipientAddr)
		})
	}
}

func Test_CurrencyIssueDecimals(t *testing.T) {
	app, server := newTestDnApp()
	defer app.CloseConnections()
	defer server.Stop()

	genCoins, err := sdk.ParseCoins("1000000000000000" + config.MainDenom)
	require.NoError(t, err)
	genAccs, _, _, genPrivKeys := CreateGenAccounts(10, genCoins)

	_, err = setGenesis(t, app, genAccs)
	require.NoError(t, err)

	recipientIdx, recipientAddr, recipientPrivKey := uint(0), genAccs[0].Address, genPrivKeys[0]
	curAmount, curDecimals, denom := sdk.OneInt(), int8(1), currency1Symbol

	// check amount with decimals currency issue
	{
		msgId, issueId := "1", "issue1"

		issueCurrency(t, app, denom, curAmount, curDecimals, msgId, issueId, recipientIdx, genAccs, genPrivKeys, true)
		checkIssueExists(t, app, issueId, denom, curAmount, recipientAddr)
		checkCurrencyExists(t, app, denom, curAmount, curDecimals)
		checkRecipientCoins(t, app, recipientAddr, denom, curAmount, curDecimals)
	}

	// check amount increase with decimals currency issue
	{
		msgId, issueId := "2", "issue2"

		newAmount := sdk.OneInt()
		curAmount = curAmount.Add(newAmount)

		issueCurrency(t, app, denom, newAmount, curDecimals, msgId, issueId, recipientIdx, genAccs, genPrivKeys, true)
		checkIssueExists(t, app, issueId, denom, newAmount, recipientAddr)
		checkCurrencyExists(t, app, denom, curAmount, curDecimals)
		checkRecipientCoins(t, app, recipientAddr, denom, curAmount, curDecimals)
	}

	// check currency issue with wrong decimals
	{
		msgId, issueId := "non-existing-msgID", "non-existing-issue"

		newAmount, newDecimals := sdk.OneInt(), curDecimals+1

		res := issueCurrency(t, app, denom, newAmount, newDecimals, msgId, issueId, recipientIdx, genAccs, genPrivKeys, false)
		CheckResultError(t, ccTypes.ErrIncorrectDecimals(0, 0, ""), res)
	}

	// check amount decrease with decimals currency issue
	{
		newAmount := sdk.OneInt()
		curAmount = curAmount.Sub(newAmount)

		destroyCurrency(t, app, chainID, denom, newAmount, recipientAddr, recipientPrivKey, true)
		checkCurrencyExists(t, app, denom, curAmount, curDecimals)
		checkRecipientCoins(t, app, recipientAddr, denom, curAmount, curDecimals)
	}

	// check currency with decimals destroy over the limit
	{
		newAmount := curAmount.Add(sdk.OneInt())

		res := destroyCurrency(t, app, chainID, denom, newAmount, recipientAddr, recipientPrivKey, false)
		CheckResultError(t, sdk.ErrInsufficientCoins(""), res)
	}
}

func Test_CurrencyDestroy(t *testing.T) {
	app, server := newTestDnApp()
	defer app.CloseConnections()
	defer server.Stop()

	genCoins, err := sdk.ParseCoins("1000000000000000" + config.MainDenom)
	require.NoError(t, err)
	genAccs, _, _, genPrivKeys := CreateGenAccounts(10, genCoins)

	_, err = setGenesis(t, app, genAccs)
	require.NoError(t, err)

	recipientIdx, recipientAddr, recipientPrivKey := uint(0), genAccs[0].Address, genPrivKeys[0]
	curSupply, denom := amount.Mul(sdk.NewInt(2)), currency1Symbol

	// check currency is issued
	{
		issueCurrency(t, app, denom, curSupply, 0, "1", issue1ID, recipientIdx, genAccs, genPrivKeys, true)
		checkIssueExists(t, app, issue1ID, denom, curSupply, recipientAddr)
		checkCurrencyExists(t, app, denom, curSupply, 0)
		checkRecipientCoins(t, app, recipientAddr, denom, curSupply, 0)
	}

	// check currency supply reduced
	{
		curSupply = curSupply.Sub(amount)
		destroyCurrency(t, app, chainID, denom, amount, recipientAddr, recipientPrivKey, true)
		checkCurrencyExists(t, app, denom, curSupply, 0)
		checkRecipientCoins(t, app, recipientAddr, denom, curSupply, 0)
	}

	// check currency destroyed
	{
		curSupply = curSupply.Sub(amount)
		require.True(t, curSupply.IsZero())

		destroyCurrency(t, app, chainID, denom, amount, recipientAddr, recipientPrivKey, true)
		checkCurrencyExists(t, app, denom, curSupply, 0)
		checkRecipientCoins(t, app, recipientAddr, denom, curSupply, 0)
	}

	// check currency destroy over the limit
	{
		res := destroyCurrency(t, app, chainID, denom, sdk.OneInt(), recipientAddr, recipientPrivKey, false)
		CheckResultError(t, sdk.ErrInsufficientCoins(""), res)
	}

	// check currency destroy with denom account doesn't have
	{
		wrongDenom := currency2Symbol

		res := destroyCurrency(t, app, chainID, wrongDenom, amount, recipientAddr, recipientPrivKey, false)
		CheckResultError(t, sdk.ErrInsufficientCoins(""), res)
	}
}

func issueCurrency(t *testing.T, app *DnServiceApp,
	ccSymbol string, ccAmount sdk.Int, ccDecimals int8, msgID, issueID string,
	recipientAccIdx uint, accs []*auth.BaseAccount, privKeys []crypto.PrivKey, doCheck bool) sdk.Result {

	issueMsg := ccMsgs.NewMsgIssueCurrency(ccSymbol, ccAmount, ccDecimals, accs[recipientAccIdx].Address, issueID)
	return MSMsgSubmitAndVote(t, app, msgID, issueMsg, recipientAccIdx, accs, privKeys, doCheck)
}

func destroyCurrency(t *testing.T, app *DnServiceApp,
	chainID, ccSymbol string, ccAmount sdk.Int,
	recipientAddr sdk.AccAddress, recipientPrivKey crypto.PrivKey, doCheck bool) sdk.Result {

	recipientAcc := GetAccountCheckTx(app, recipientAddr)
	destroyMsg := ccMsgs.NewMsgDestroyCurrency(chainID, ccSymbol, ccAmount, recipientAcc.GetAddress(), recipientAcc.GetAddress().String())
	tx := genTx([]sdk.Msg{destroyMsg}, []uint64{recipientAcc.GetAccountNumber()}, []uint64{recipientAcc.GetSequence()}, recipientPrivKey)

	res := DeliverTx(app, tx)
	if doCheck {
		require.True(t, res.IsOK())
	}

	return res
}

func checkCurrencyExists(t *testing.T, app *DnServiceApp, symbol string, supply sdk.Int, decimals int8) {
	currencyObj := ccTypes.Currency{}
	CheckRunQuery(t, app, ccTypes.CurrencyReq{Symbol: symbol}, queryCurrencyGetCurrencyPath, &currencyObj)

	require.Equal(t, symbol, currencyObj.Symbol, "symbol")
	require.True(t, currencyObj.Supply.Equal(supply), "supply")
	require.Equal(t, decimals, currencyObj.Decimals, "decimals")
}

func checkIssueExists(t *testing.T, app *DnServiceApp, issueID, symbol string, amount sdk.Int, recipientAddr sdk.AccAddress) {
	issue := ccTypes.Issue{}
	CheckRunQuery(t, app, ccTypes.IssueReq{IssueID: issueID}, queryCurrencyGetIssuePath, &issue)

	require.Equal(t, symbol, issue.Symbol, "symbol")
	require.True(t, issue.Amount.Equal(amount), "amount")
	require.Equal(t, recipientAddr, issue.Recipient)
}

func checkRecipientCoins(t *testing.T, app *DnServiceApp, recipientAddr sdk.AccAddress, denom string, amount sdk.Int, decimals int8) {
	checkBalance := amount

	coins := app.bankKeeper.GetCoins(GetContext(app, true), recipientAddr)
	actualBalance := coins.AmountOf(denom)

	require.True(t, actualBalance.Equal(checkBalance), " denom %q, checkBalance / actualBalance mismatch: %s / %s", denom, checkBalance.String(), actualBalance.String())
}
