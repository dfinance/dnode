// +build unit

package app

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"

	dnTypes "github.com/dfinance/dnode/helpers/types"
	ccTypes "github.com/dfinance/dnode/x/currencies"
	msTypes "github.com/dfinance/dnode/x/multisig/types"
)

const (
	queryCurrencyIssuePath    = "/custom/" + ccTypes.ModuleName + "/" + ccTypes.QueryIssue
	queryCurrencyCurrencyPath = "/custom/" + ccTypes.ModuleName + "/" + ccTypes.QueryCurrency
	queryCurrencyDestroyPath  = "/custom/" + ccTypes.ModuleName + "/" + ccTypes.QueryDestroy
	queryCurrencyDestroysPath = "/custom/" + ccTypes.ModuleName + "/" + ccTypes.QueryDestroys
)

// Checks that currencies module supports only multisig calls for issue msg (using MSRouter).
func TestCurrenciesApp_MultisigHandler(t *testing.T) {
	t.Parallel()
	app, server := newTestDnApp()
	defer app.CloseConnections()
	defer server.Stop()

	genValidators, _, _, genPrivKeys := CreateGenAccounts(7, GenDefCoins(t))

	_, err := setGenesis(t, app, genValidators)
	require.NoError(t, err)

	{
		senderAcc, senderPrivKey := GetAccountCheckTx(app, genValidators[0].Address), genPrivKeys[0]
		issueMsg := ccTypes.NewMsgIssueCurrency(issue1ID, currency1Denom, amount, 0, senderAcc.GetAddress())
		tx := genTx([]sdk.Msg{issueMsg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
		CheckDeliverSpecificErrorTx(t, app, tx, sdkErrors.ErrUnauthorized)
	}
}

// Test currencies module queries.
func TestCurrenciesApp_Queries(t *testing.T) {
	t.Parallel()
	app, server := newTestDnApp()
	defer app.CloseConnections()
	defer server.Stop()

	genAccs, _, _, genPrivKeys := CreateGenAccounts(10, GenDefCoins(t))

	_, err := setGenesis(t, app, genAccs)
	require.NoError(t, err)

	recipientIdx, recipientAddr, recipientPrivKey := uint(0), genAccs[0].Address, genPrivKeys[0]

	checkDestroyQueryObj := func(obj ccTypes.Destroy, id uint64, denom string, amount sdk.Int, spenderAddr sdk.AccAddress) {
		require.Equal(t, id, obj.ID.UInt64())
		require.Equal(t, denom, obj.Denom)
		require.True(t, obj.Amount.Equal(amount))
		require.Equal(t, spenderAddr, obj.Spender)
		require.Equal(t, chainID, obj.ChainID)
	}

	// issue multiple currencies
	issueCurrency(t, app, currency1Denom, amount, 0, "msg1", issue1ID, recipientIdx, genAccs, genPrivKeys, true)
	issueCurrency(t, app, currency2Denom, amount, 0, "msg2", issue2ID, recipientIdx, genAccs, genPrivKeys, true)
	issueCurrency(t, app, currency3Denom, amount, 0, "msg3", issue3ID, recipientIdx, genAccs, genPrivKeys, true)

	// check getCurrency query
	{
		checkCurrencyExists(t, app, currency1Denom, amount, 0)
		checkCurrencyExists(t, app, currency2Denom, amount, 0)
		checkCurrencyExists(t, app, currency3Denom, amount, 0)
	}

	// check getIssue query
	{
		checkIssueExists(t, app, issue1ID, currency1Denom, amount, recipientAddr)
		checkIssueExists(t, app, issue2ID, currency2Denom, amount, recipientAddr)
		checkIssueExists(t, app, issue3ID, currency3Denom, amount, recipientAddr)
	}

	// destroy currencies
	destroyAmount := amount.QuoRaw(3)
	destroyCurrency(t, app, chainID, currency3Denom, destroyAmount, recipientAddr, recipientPrivKey, true)
	destroyCurrency(t, app, chainID, currency3Denom, destroyAmount, recipientAddr, recipientPrivKey, true)
	destroyCurrency(t, app, chainID, currency3Denom, destroyAmount, recipientAddr, recipientPrivKey, true)

	// check getDestroys query with pagination
	{
		// page 1
		{
			destroys := ccTypes.Destroys{}
			reqParams := ccTypes.DestroysReq{Page: sdk.NewUint(1), Limit: sdk.NewUint(2)}
			CheckRunQuery(t, app, reqParams, queryCurrencyDestroysPath, &destroys)

			require.Len(t, destroys, 2)
			checkDestroyQueryObj(destroys[0], 0, currency3Denom, destroyAmount, recipientAddr)
			checkDestroyQueryObj(destroys[1], 1, currency3Denom, destroyAmount, recipientAddr)
		}

		// page 2
		{
			destroys := ccTypes.Destroys{}
			reqParams := ccTypes.DestroysReq{Page: sdk.NewUint(2), Limit: sdk.NewUint(2)}
			CheckRunQuery(t, app, reqParams, queryCurrencyDestroysPath, &destroys)

			require.Len(t, destroys, 1)
			checkDestroyQueryObj(destroys[0], 2, currency3Denom, destroyAmount, recipientAddr)
		}
	}

	// check getDestroy query
	{
		checkDestroyExists(t, app, 0, currency3Denom, destroyAmount, recipientAddr, recipientAddr.String())
		checkDestroyExists(t, app, 1, currency3Denom, destroyAmount, recipientAddr, recipientAddr.String())
		checkDestroyExists(t, app, 2, currency3Denom, destroyAmount, recipientAddr, recipientAddr.String())
	}
}

// Test currency issue logic with failure scenarios.
func TestCurrenciesApp_Issue(t *testing.T) {
	t.Parallel()
	app, server := newTestDnApp()
	defer app.CloseConnections()
	defer server.Stop()

	genAccs, _, _, genPrivKeys := CreateGenAccounts(10, GenDefCoins(t))

	_, err := setGenesis(t, app, genAccs)
	require.NoError(t, err)

	recipientIdx, recipientAddr := uint(0), genAccs[0].Address
	curAmount, curDecimals, denom := amount, uint8(0), currency1Denom

	// ok: currency is issued
	{
		msgId, issueId := "1", "issue1"

		issueCurrency(t, app, denom, curAmount, curDecimals, msgId, issueId, recipientIdx, genAccs, genPrivKeys, true)
		checkIssueExists(t, app, issueId, denom, curAmount, recipientAddr)
		checkCurrencyExists(t, app, denom, curAmount, curDecimals)
		checkRecipientCoins(t, app, recipientAddr, denom, curAmount, curDecimals)
	}

	// ok currency supply increased
	{
		msgId, issueId := "2", "issue2"
		newAmount := sdk.NewInt(200)
		curAmount = curAmount.Add(newAmount)

		issueCurrency(t, app, denom, newAmount, curDecimals, msgId, issueId, recipientIdx, genAccs, genPrivKeys, true)
		checkIssueExists(t, app, issueId, denom, newAmount, recipientAddr)
		checkCurrencyExists(t, app, denom, curAmount, curDecimals)
		checkRecipientCoins(t, app, recipientAddr, denom, curAmount, curDecimals)
	}

	// fail: currency issue for existing currency with different decimals
	{
		msgId, issueId := "3", "issue3"

		res, err := issueCurrency(t, app, denom, sdk.OneInt(), curDecimals+1, msgId, issueId, recipientIdx, genAccs, genPrivKeys, false)
		CheckResultError(t, ccTypes.ErrIncorrectDecimals, res, err)
	}

	// fail: currency issue with the same issueID
	{
		msgId, issueId := "non-existing-msgID", "issue1"

		res, err := issueCurrency(t, app, denom, amount, 0, msgId, issueId, recipientIdx, genAccs, genPrivKeys, false)
		CheckResultError(t, ccTypes.ErrWrongIssueID, res, err)
	}

	// fail: currency issue with already existing uniqueMsgID
	{
		msgId, issueId := "1", "non-existing-issue"

		res, err := issueCurrency(t, app, denom, amount, 0, msgId, issueId, recipientIdx, genAccs, genPrivKeys, false)
		CheckResultError(t, msTypes.ErrNotUniqueID, res, err)
	}
}

// Test maximum bank supply level (DVM has u128 limit).
func TestCurrenciesApp_IssueHugeAmount(t *testing.T) {
	t.Parallel()
	app, server := newTestDnApp()
	defer app.CloseConnections()
	defer server.Stop()

	genAccs, _, _, genPrivKeys := CreateGenAccounts(10, GenDefCoins(t))

	_, err := setGenesis(t, app, genAccs)
	require.NoError(t, err)

	recipientIdx, recipientAddr := uint(0), genAccs[0].Address

	// check huge amount currency issue (max value for u128)
	{
		msgId, issueId, denom := "1", "issue1", currency1Denom

		hugeAmount, ok := sdk.NewIntFromString("100000000000000000000000000000000000000")
		require.True(t, ok)

		issueCurrency(t, app, denom, hugeAmount, 0, msgId, issueId, recipientIdx, genAccs, genPrivKeys, true)
		checkIssueExists(t, app, issueId, denom, hugeAmount, recipientAddr)
		checkCurrencyExists(t, app, denom, hugeAmount, 0)
		checkRecipientCoins(t, app, recipientAddr, denom, hugeAmount, 0)
	}

	// check huge amount currency issue (that worked before u128)
	{
		msgId, issueId, denom := "2", "issue2", currency2Denom

		hugeAmount, ok := sdk.NewIntFromString("1000000000000000000000000000000000000000000000")
		require.True(t, ok)

		issueCurrency(t, app, denom, hugeAmount, 0, msgId, issueId, recipientIdx, genAccs, genPrivKeys, true)
		checkIssueExists(t, app, issueId, denom, hugeAmount, recipientAddr)
		checkCurrencyExists(t, app, denom, hugeAmount, 0)

		require.Panics(t, func() {
			app.bankKeeper.GetCoins(GetContext(app, true), recipientAddr)
		})
	}
}

// Test issue/destroy currency with decimals.
func TestCurrenciesApp_Decimals(t *testing.T) {
	t.Parallel()
	app, server := newTestDnApp()
	defer app.CloseConnections()
	defer server.Stop()

	genAccs, _, _, genPrivKeys := CreateGenAccounts(10, GenDefCoins(t))

	_, err := setGenesis(t, app, genAccs)
	require.NoError(t, err)

	recipientIdx, recipientAddr, recipientPrivKey := uint(0), genAccs[0].Address, genPrivKeys[0]
	curAmount, curDecimals, denom := sdk.OneInt(), uint8(1), currency1Denom

	// issue currency amount with decimals
	{
		msgId, issueId := "1", "issue1"

		issueCurrency(t, app, denom, curAmount, curDecimals, msgId, issueId, recipientIdx, genAccs, genPrivKeys, true)
		checkIssueExists(t, app, issueId, denom, curAmount, recipientAddr)
		checkCurrencyExists(t, app, denom, curAmount, curDecimals)
		checkRecipientCoins(t, app, recipientAddr, denom, curAmount, curDecimals)
	}

	// increase currency amount with decimals
	{
		msgId, issueId := "2", "issue2"

		newAmount := sdk.OneInt()
		curAmount = curAmount.Add(newAmount)

		issueCurrency(t, app, denom, newAmount, curDecimals, msgId, issueId, recipientIdx, genAccs, genPrivKeys, true)
		checkIssueExists(t, app, issueId, denom, newAmount, recipientAddr)
		checkCurrencyExists(t, app, denom, curAmount, curDecimals)
		checkRecipientCoins(t, app, recipientAddr, denom, curAmount, curDecimals)
	}

	// decrease currency amount with decimals
	{
		newAmount := sdk.OneInt()
		curAmount = curAmount.Sub(newAmount)

		destroyCurrency(t, app, chainID, denom, newAmount, recipientAddr, recipientPrivKey, true)
		checkCurrencyExists(t, app, denom, curAmount, curDecimals)
		checkRecipientCoins(t, app, recipientAddr, denom, curAmount, curDecimals)
	}
}

// Test destroy currency with fail scenarios.
func TestCurrenciesApp_Destroy(t *testing.T) {
	t.Parallel()
	app, server := newTestDnApp()
	defer app.CloseConnections()
	defer server.Stop()

	genAccs, _, _, genPrivKeys := CreateGenAccounts(10, GenDefCoins(t))

	_, err := setGenesis(t, app, genAccs)
	require.NoError(t, err)

	recipientIdx, recipientAddr, recipientPrivKey := uint(0), genAccs[0].Address, genPrivKeys[0]
	curSupply, denom := amount.Mul(sdk.NewInt(2)), currency1Denom

	// issue currency
	{
		issueCurrency(t, app, denom, curSupply, 0, "1", issue1ID, recipientIdx, genAccs, genPrivKeys, true)
		checkIssueExists(t, app, issue1ID, denom, curSupply, recipientAddr)
		checkCurrencyExists(t, app, denom, curSupply, 0)
		checkRecipientCoins(t, app, recipientAddr, denom, curSupply, 0)
	}

	// ok: destroy currency
	{
		curSupply = curSupply.Sub(amount)
		destroyCurrency(t, app, chainID, denom, amount, recipientAddr, recipientPrivKey, true)
		checkDestroyExists(t, app, 0, denom, amount, recipientAddr, recipientAddr.String())
		checkCurrencyExists(t, app, denom, curSupply, 0)
		checkRecipientCoins(t, app, recipientAddr, denom, curSupply, 0)
	}

	// ok: destroy currency (currency supply is 0)
	{
		curSupply = curSupply.Sub(amount)
		require.True(t, curSupply.IsZero())

		destroyCurrency(t, app, chainID, denom, amount, recipientAddr, recipientPrivKey, true)
		checkDestroyExists(t, app, 1, denom, amount, recipientAddr, recipientAddr.String())
		checkCurrencyExists(t, app, denom, curSupply, 0)
		checkRecipientCoins(t, app, recipientAddr, denom, curSupply, 0)
	}

	// fail: currency destroy over the limit
	{
		res, err := destroyCurrency(t, app, chainID, denom, sdk.OneInt(), recipientAddr, recipientPrivKey, false)
		CheckResultError(t, sdkErrors.ErrInsufficientFunds, res, err)
	}

	// fail: currency destroy with denom account doesn't have
	{
		wrongDenom := currency2Denom

		res, err := destroyCurrency(t, app, chainID, wrongDenom, amount, recipientAddr, recipientPrivKey, false)
		CheckResultError(t, sdkErrors.ErrInsufficientFunds, res, err)
	}
}

func issueCurrency(t *testing.T, app *DnServiceApp,
	ccDenom string, ccAmount sdk.Int, ccDecimals uint8, msgID, issueID string,
	recipientAccIdx uint, accs []*auth.BaseAccount, privKeys []crypto.PrivKey, doCheck bool) (*sdk.Result, error) {

	issueMsg := ccTypes.NewMsgIssueCurrency(issueID, ccDenom, ccAmount, ccDecimals, accs[recipientAccIdx].Address)
	return MSMsgSubmitAndVote(t, app, msgID, issueMsg, recipientAccIdx, accs, privKeys, doCheck)
}

func destroyCurrency(t *testing.T, app *DnServiceApp,
	chainID, ccDenom string, ccAmount sdk.Int,
	recipientAddr sdk.AccAddress, recipientPrivKey crypto.PrivKey, doCheck bool) (*sdk.Result, error) {

	recipientAcc := GetAccountCheckTx(app, recipientAddr)
	destroyMsg := ccTypes.NewMsgDestroyCurrency(ccDenom, ccAmount, recipientAcc.GetAddress(), recipientAcc.GetAddress().String(), chainID)
	tx := genTx([]sdk.Msg{destroyMsg}, []uint64{recipientAcc.GetAccountNumber()}, []uint64{recipientAcc.GetSequence()}, recipientPrivKey)

	res, err := DeliverTx(app, tx)
	if doCheck {
		require.NoError(t, err)
	}

	return res, err
}

func checkCurrencyExists(t *testing.T, app *DnServiceApp, denom string, supply sdk.Int, decimals uint8) {
	currencyObj := ccTypes.Currency{}
	CheckRunQuery(t, app, ccTypes.CurrencyReq{Denom: denom}, queryCurrencyCurrencyPath, &currencyObj)

	require.Equal(t, denom, currencyObj.Denom, "denom")
	require.Equal(t, decimals, currencyObj.Decimals, "decimals")
	require.True(t, currencyObj.Supply.Equal(supply), "supply")
}

func checkIssueExists(t *testing.T, app *DnServiceApp, issueID, denom string, amount sdk.Int, payeeAddr sdk.AccAddress) {
	issue := ccTypes.Issue{}
	CheckRunQuery(t, app, ccTypes.IssueReq{ID: issueID}, queryCurrencyIssuePath, &issue)

	require.Equal(t, denom, issue.Denom, "symbol")
	require.True(t, issue.Amount.Equal(amount), "amount")
	require.Equal(t, payeeAddr, issue.Payee)
}

func checkDestroyExists(t *testing.T, app *DnServiceApp, id uint64, denom string, amount sdk.Int, spenderAddr sdk.AccAddress, recipient string) {
	destroy := ccTypes.Destroy{}
	CheckRunQuery(t, app, ccTypes.DestroyReq{ID: dnTypes.NewIDFromUint64(id)}, queryCurrencyDestroyPath, &destroy)

	require.Equal(t, id, destroy.ID.UInt64())
	require.Equal(t, denom, destroy.Denom)
	require.True(t, destroy.Amount.Equal(amount))
	require.Equal(t, spenderAddr, destroy.Spender)
	require.Equal(t, recipient, destroy.Recipient)
	require.Equal(t, chainID, destroy.ChainID)
}

func checkRecipientCoins(t *testing.T, app *DnServiceApp, recipientAddr sdk.AccAddress, denom string, amount sdk.Int, decimals uint8) {
	checkBalance := amount

	coins := app.bankKeeper.GetCoins(GetContext(app, true), recipientAddr)
	actualBalance := coins.AmountOf(denom)

	require.True(t, actualBalance.Equal(checkBalance), " denom %q, checkBalance / actualBalance mismatch: %s / %s", denom, checkBalance.String(), actualBalance.String())
}
