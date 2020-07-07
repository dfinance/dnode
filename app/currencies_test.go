// +build unit

package app

import (
	"fmt"
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authExported "github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/supply"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"

	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/ccstorage"
	"github.com/dfinance/dnode/x/currencies"
	"github.com/dfinance/dnode/x/multisig"
)

const (
	queryCurrencyIssuePath     = "/custom/" + currencies.ModuleName + "/" + currencies.QueryIssue
	queryCurrencyCurrencyPath  = "/custom/" + currencies.ModuleName + "/" + currencies.QueryCurrency
	queryCurrencyWithdrawsPath = "/custom/" + currencies.ModuleName + "/" + currencies.QueryWithdraws
	queryCurrencyWithdrawPath  = "/custom/" + currencies.ModuleName + "/" + currencies.QueryWithdraw
)

type CalculatedSupply struct {
	Denom    string
	Supply   sdk.Int
	Accounts []AccountBalance
}

type AccountBalance struct {
	Name   string
	Amount sdk.Int
}

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
		issueMsg := currencies.NewMsgIssueCurrency(issue1ID, currency1Denom, amount, 0, senderAcc.GetAddress())
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

	checkWithdrawQueryObj := func(obj currencies.Withdraw, id uint64, denom string, amount sdk.Int, spenderAddr sdk.AccAddress) {
		require.Equal(t, id, obj.ID.UInt64())
		require.Equal(t, denom, obj.Denom)
		require.True(t, obj.Amount.Equal(amount))
		require.Equal(t, spenderAddr, obj.Spender)
		require.Equal(t, chainID, obj.PegZoneChainID)
	}

	// issue multiple currencies
	createCurrency(t, app, currency1Denom, 0)
	createCurrency(t, app, currency2Denom, 0)
	createCurrency(t, app, currency3Denom, 0)
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

	// withdraw currencies
	withdrawAmount := amount.QuoRaw(3)
	withdrawCurrency(t, app, chainID, currency3Denom, withdrawAmount, recipientAddr, recipientPrivKey, true)
	withdrawCurrency(t, app, chainID, currency3Denom, withdrawAmount, recipientAddr, recipientPrivKey, true)
	withdrawCurrency(t, app, chainID, currency3Denom, withdrawAmount, recipientAddr, recipientPrivKey, true)

	// check getWithdraws query with pagination
	{
		// page 1
		{
			withdraws := currencies.Withdraws{}
			reqParams := currencies.WithdrawsReq{Page: sdk.NewUint(1), Limit: sdk.NewUint(2)}
			CheckRunQuery(t, app, reqParams, queryCurrencyWithdrawsPath, &withdraws)

			require.Len(t, withdraws, 2)
			checkWithdrawQueryObj(withdraws[0], 0, currency3Denom, withdrawAmount, recipientAddr)
			checkWithdrawQueryObj(withdraws[1], 1, currency3Denom, withdrawAmount, recipientAddr)
		}

		// page 2
		{
			withdraws := currencies.Withdraws{}
			reqParams := currencies.WithdrawsReq{Page: sdk.NewUint(2), Limit: sdk.NewUint(2)}
			CheckRunQuery(t, app, reqParams, queryCurrencyWithdrawsPath, &withdraws)

			require.Len(t, withdraws, 1)
			checkWithdrawQueryObj(withdraws[0], 2, currency3Denom, withdrawAmount, recipientAddr)
		}
	}

	// check getWithdraw query
	{
		checkWithdrawExists(t, app, 0, currency3Denom, withdrawAmount, recipientAddr, recipientAddr.String())
		checkWithdrawExists(t, app, 1, currency3Denom, withdrawAmount, recipientAddr, recipientAddr.String())
		checkWithdrawExists(t, app, 2, currency3Denom, withdrawAmount, recipientAddr, recipientAddr.String())
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

	createCurrency(t, app, denom, curDecimals)

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
		CheckResultError(t, currencies.ErrIncorrectDecimals, res, err)
	}

	// fail: currency issue with the same issueID
	{
		msgId, issueId := "non-existing-msgID", "issue1"

		res, err := issueCurrency(t, app, denom, amount, 0, msgId, issueId, recipientIdx, genAccs, genPrivKeys, false)
		CheckResultError(t, currencies.ErrWrongIssueID, res, err)
	}

	// fail: currency issue with already existing uniqueMsgID
	{
		msgId, issueId := "1", "non-existing-issue"

		res, err := issueCurrency(t, app, denom, amount, 0, msgId, issueId, recipientIdx, genAccs, genPrivKeys, false)
		CheckResultError(t, multisig.ErrWrongCallUniqueId, res, err)
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

		createCurrency(t, app, denom, 0)
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

		createCurrency(t, app, denom, 0)
		issueCurrency(t, app, denom, hugeAmount, 0, msgId, issueId, recipientIdx, genAccs, genPrivKeys, true)
		checkIssueExists(t, app, issueId, denom, hugeAmount, recipientAddr)
		checkCurrencyExists(t, app, denom, hugeAmount, 0)

		require.Panics(t, func() {
			app.bankKeeper.GetCoins(GetContext(app, true), recipientAddr)
		})
	}
}

// Test issue/withdraw currency with decimals.
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

	createCurrency(t, app, denom, curDecimals)

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

		withdrawCurrency(t, app, chainID, denom, newAmount, recipientAddr, recipientPrivKey, true)
		checkCurrencyExists(t, app, denom, curAmount, curDecimals)
		checkRecipientCoins(t, app, recipientAddr, denom, curAmount, curDecimals)
	}
}

// Test withdraw currency with fail scenarios.
func TestCurrenciesApp_Withdraw(t *testing.T) {
	t.Parallel()

	app, server := newTestDnApp()
	defer app.CloseConnections()
	defer server.Stop()

	genAccs, _, _, genPrivKeys := CreateGenAccounts(10, GenDefCoins(t))

	_, err := setGenesis(t, app, genAccs)
	require.NoError(t, err)

	recipientIdx, recipientAddr, recipientPrivKey := uint(0), genAccs[0].Address, genPrivKeys[0]
	curSupply, denom := amount.Mul(sdk.NewInt(2)), currency1Denom

	createCurrency(t, app, denom, 0)

	// issue currency
	{
		issueCurrency(t, app, denom, curSupply, 0, "1", issue1ID, recipientIdx, genAccs, genPrivKeys, true)
		checkIssueExists(t, app, issue1ID, denom, curSupply, recipientAddr)
		checkCurrencyExists(t, app, denom, curSupply, 0)
		checkRecipientCoins(t, app, recipientAddr, denom, curSupply, 0)
	}

	// ok: withdraw currency
	{
		curSupply = curSupply.Sub(amount)
		withdrawCurrency(t, app, chainID, denom, amount, recipientAddr, recipientPrivKey, true)
		checkWithdrawExists(t, app, 0, denom, amount, recipientAddr, recipientAddr.String())
		checkCurrencyExists(t, app, denom, curSupply, 0)
		checkRecipientCoins(t, app, recipientAddr, denom, curSupply, 0)
	}

	// ok: withdraw currency (currency supply is 0)
	{
		curSupply = curSupply.Sub(amount)
		require.True(t, curSupply.IsZero())

		withdrawCurrency(t, app, chainID, denom, amount, recipientAddr, recipientPrivKey, true)
		checkWithdrawExists(t, app, 1, denom, amount, recipientAddr, recipientAddr.String())
		checkCurrencyExists(t, app, denom, curSupply, 0)
		checkRecipientCoins(t, app, recipientAddr, denom, curSupply, 0)
	}

	// fail: currency withdraw over the limit
	{
		res, err := withdrawCurrency(t, app, chainID, denom, sdk.OneInt(), recipientAddr, recipientPrivKey, false)
		CheckResultError(t, sdkErrors.ErrInsufficientFunds, res, err)
	}

	// fail: currency withdraw with denom account doesn't have
	{
		wrongDenom := currency2Denom

		res, err := withdrawCurrency(t, app, chainID, wrongDenom, amount, recipientAddr, recipientPrivKey, false)
		CheckResultError(t, ccstorage.ErrWrongDenom, res, err)
	}
}

// Test issues and destroys currency and verifies that supply (via supply module) stays up-to-date.
func TestCurrenciesApp_Supply(t *testing.T) {
	t.Parallel()

	app, server := newTestDnApp()
	defer app.CloseConnections()
	defer server.Stop()

	genAccs, _, _, genPrivKeys := CreateGenAccounts(10, GenDefCoins(t))

	_, err := setGenesis(t, app, genAccs)
	require.NoError(t, err)

	getCalculatedSupplies := func() map[string]CalculatedSupply {
		supplies := make(map[string]CalculatedSupply, 0)
		app.accountKeeper.IterateAccounts(GetContext(app, true), func(acc authExported.Account) bool {
			accName := ""
			if modAcc, ok := acc.(*supply.ModuleAccount); ok {
				accName = modAcc.GetName()
			} else {
				accName = acc.GetAddress().String()
			}

			for _, coin := range acc.GetCoins() {
				denomSupply, ok := supplies[coin.Denom]
				if !ok {
					denomSupply = CalculatedSupply{
						Denom:    coin.Denom,
						Supply:   sdk.ZeroInt(),
						Accounts: make([]AccountBalance, 0),
					}
				}

				accBalance := AccountBalance{Name: accName, Amount: coin.Amount}

				denomSupply.Supply = denomSupply.Supply.Add(coin.Amount)
				denomSupply.Accounts = append(denomSupply.Accounts, accBalance)

				supplies[coin.Denom] = denomSupply
			}

			return false
		})
		return supplies
	}

	getModuleSupplies := func() map[string]sdk.Int {
		supplies := make(map[string]sdk.Int, 0)
		for _, coin := range app.supplyKeeper.GetSupply(GetContext(app, true)).GetTotal() {
			supplies[coin.Denom] = coin.Amount
		}
		return supplies
	}

	getCCSupplies := func() map[string]sdk.Int {
		supplies := make(map[string]sdk.Int, 0)
		for denom := range app.ccsKeeper.GetCurrenciesParams(GetContext(app, true)) {
			currency, err := app.ccsKeeper.GetCurrency(GetContext(app, true), denom)
			require.NoError(t, err, "requesting ccStorage for %q currency", denom)

			supplies[denom] = currency.Supply
		}
		return supplies
	}

	checkSupplies := func(testID string) {
		calcSupplies := getCalculatedSupplies()
		modSupplies := getModuleSupplies()
		ccSupplies := getCCSupplies()

		for denom, ccSupply := range ccSupplies {
			errBuilder := strings.Builder{}

			modSupply, modFound := modSupplies[denom]
			if !modFound {
				modSupply = sdk.ZeroInt()
			}

			calcSupply, calcFound := calcSupplies[denom]
			if !calcFound {
				calcSupply.Supply = sdk.ZeroInt()
			}

			errBuilder.WriteString(fmt.Sprintf(">> %s: denom: %s\n", testID, denom))
			errBuilder.WriteString(fmt.Sprintf("\tmod supply:  %s\n", modSupply.String()))
			errBuilder.WriteString(fmt.Sprintf("\tccs supply:  %s\n", ccSupply.String()))
			errBuilder.WriteString(fmt.Sprintf("\tcalc supply: %s\n", calcSupply.Supply.String()))

			for _, acc := range calcSupply.Accounts {
				errBuilder.WriteString(fmt.Sprintf("\t-%s: %s\n", acc.Name, acc.Amount.String()))
			}

			allEqual := false
			if modSupply.Equal(calcSupply.Supply) && modSupply.Equal(ccSupply) {
				allEqual = true
				errBuilder.WriteString("\t-> OK\n")
			} else {
				errBuilder.WriteString(fmt.Sprintf("\t-> Not equal, mod/calc diff: %s\n", modSupply.Sub(calcSupply.Supply).String()))
				errBuilder.WriteString(fmt.Sprintf("\t-> Not equal: mod/ccs diff:  %s\n", modSupply.Sub(ccSupply).String()))
			}

			require.True(t, allEqual, errBuilder.String())
		}
	}

	// initial check
	{
		checkSupplies("initial")
	}

	// issue 50.0 dfi to account1
	{
		amount, _ := sdk.NewIntFromString("50000000000000000000")
		issueCurrency(t, app, "dfi", amount, 18, "1", issue1ID, uint(0), genAccs, genPrivKeys, true)

		checkSupplies("50.0 dfi issued to acc #1")
	}

	// issue 5.0 btc to account2
	{
		amount, _ := sdk.NewIntFromString("500000000")
		issueCurrency(t, app, "btc", amount, 8, "2", issue2ID, uint(1), genAccs, genPrivKeys, true)

		checkSupplies("5.0 btc issued to acc #2")
	}

	// withdraw 2.5 btc from account2
	{
		recipientAddr, recipientPrivKey := genAccs[1].Address, genPrivKeys[1]
		amount, _ := sdk.NewIntFromString("250000000")
		withdrawCurrency(t, app, chainID, "btc", amount, recipientAddr, recipientPrivKey, true)

		checkSupplies("2.5 btc destroyed from acc #2")
	}

	// transfer 1.0 btc from account2 to account1
	{
		coin := sdk.NewCoin("btc", sdk.NewInt(100000000))
		coins := sdk.NewCoins(coin)
		payerAddr, payeeAddr := genAccs[1].Address, genAccs[0].Address

		app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{ChainID: chainID, Height: app.LastBlockHeight() + 1}})
		ctx := GetContext(app, false)

		require.NoError(t, app.bankKeeper.SendCoins(ctx, payerAddr, payeeAddr, coins))

		app.EndBlock(abci.RequestEndBlock{})
		app.Commit()

		checkSupplies("transfer 1.0 btc from acc #2 to acc #1")
	}
}

// createCurrency creates currency with random VM paths.
func createCurrency(t *testing.T, app *DnServiceApp, ccDenom string, ccDecimals uint8) {
	_, balancePathHex := GenerateRandomBytes(10)
	_, infoPathHex := GenerateRandomBytes(10)

	params := ccstorage.CurrencyParams{
		Decimals:       ccDecimals,
		BalancePathHex: balancePathHex,
		InfoPathHex:    infoPathHex,
	}

	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{ChainID: chainID, Height: app.LastBlockHeight() + 1}})
	err := app.ccKeeper.CreateCurrency(GetContext(app, false), ccDenom, params)
	require.NoError(t, err, "creating %q currency", ccDenom)
	app.EndBlock(abci.RequestEndBlock{})
	app.Commit()
}

// issueCurrency creates currency issue multisig message and confirms it.
func issueCurrency(t *testing.T, app *DnServiceApp,
	ccDenom string, ccAmount sdk.Int, ccDecimals uint8, msgID, issueID string,
	recipientAccIdx uint, accs []*auth.BaseAccount, privKeys []crypto.PrivKey, doCheck bool) (*sdk.Result, error) {

	issueMsg := currencies.NewMsgIssueCurrency(issueID, ccDenom, ccAmount, ccDecimals, accs[recipientAccIdx].Address)
	return MSMsgSubmitAndVote(t, app, msgID, issueMsg, recipientAccIdx, accs, privKeys, doCheck)
}

// withdrawCurrency creates withdraw currency multisig message and confirms it.
func withdrawCurrency(t *testing.T, app *DnServiceApp,
	chainID, ccDenom string, ccAmount sdk.Int,
	spenderAddr sdk.AccAddress, spenderPrivKey crypto.PrivKey, doCheck bool) (*sdk.Result, error) {

	spenderAcc := GetAccountCheckTx(app, spenderAddr)
	withdrawMsg := currencies.NewMsgWithdrawCurrency(ccDenom, ccAmount, spenderAcc.GetAddress(), spenderAcc.GetAddress().String(), chainID)
	tx := genTx([]sdk.Msg{withdrawMsg}, []uint64{spenderAcc.GetAccountNumber()}, []uint64{spenderAcc.GetSequence()}, spenderPrivKey)

	res, err := DeliverTx(app, tx)
	if doCheck {
		require.NoError(t, err)
	}

	return res, err
}

// checkCurrencyExists checks currency exists.
func checkCurrencyExists(t *testing.T, app *DnServiceApp, denom string, supply sdk.Int, decimals uint8) {
	currencyObj := ccstorage.Currency{}
	CheckRunQuery(t, app, currencies.CurrencyReq{Denom: denom}, queryCurrencyCurrencyPath, &currencyObj)

	require.Equal(t, denom, currencyObj.Denom, "denom")
	require.Equal(t, decimals, currencyObj.Decimals, "decimals")
	require.True(t, currencyObj.Supply.Equal(supply), "supply")
}

// checkIssueExists checks issue exists.
func checkIssueExists(t *testing.T, app *DnServiceApp, issueID, denom string, amount sdk.Int, payeeAddr sdk.AccAddress) {
	issue := currencies.Issue{}
	CheckRunQuery(t, app, currencies.IssueReq{ID: issueID}, queryCurrencyIssuePath, &issue)

	require.Equal(t, denom, issue.Denom, "symbol")
	require.True(t, issue.Amount.Equal(amount), "amount")
	require.Equal(t, payeeAddr, issue.Payee)
}

// checkWithdrawExists checks withdraw exists.
func checkWithdrawExists(t *testing.T, app *DnServiceApp, id uint64, denom string, amount sdk.Int, spenderAddr sdk.AccAddress, pzSpender string) {
	withdraw := currencies.Withdraw{}
	CheckRunQuery(t, app, currencies.WithdrawReq{ID: dnTypes.NewIDFromUint64(id)}, queryCurrencyWithdrawPath, &withdraw)

	require.Equal(t, id, withdraw.ID.UInt64())
	require.Equal(t, denom, withdraw.Denom)
	require.True(t, withdraw.Amount.Equal(amount))
	require.Equal(t, spenderAddr, withdraw.Spender)
	require.Equal(t, pzSpender, withdraw.PegZoneSpender)
	require.Equal(t, chainID, withdraw.PegZoneChainID)
}

// checkRecipientCoins checks account balance.
func checkRecipientCoins(t *testing.T, app *DnServiceApp, recipientAddr sdk.AccAddress, denom string, amount sdk.Int, decimals uint8) {
	checkBalance := amount

	coins := app.bankKeeper.GetCoins(GetContext(app, true), recipientAddr)
	actualBalance := coins.AmountOf(denom)

	require.True(t, actualBalance.Equal(checkBalance), " denom %q, checkBalance / actualBalance mismatch: %s / %s", denom, checkBalance.String(), actualBalance.String())

	balances, err := app.ccsKeeper.GetAccountBalanceResources(GetContext(app, true), recipientAddr)
	require.NoError(t, err, "denom %q: reading balance resources", denom)
	for _, balance := range balances {
		if balance.Denom == denom {
			require.Equal(t, amount.String(), balance.Resource.Value.String(), "denom %q: checking balance resource value", denom)
		}
	}
}
