package app

import (
	"github.com/WingsDao/wings-blockchain/x/currencies"
	curMsgs "github.com/WingsDao/wings-blockchain/x/currencies/msgs"
	curTypes "github.com/WingsDao/wings-blockchain/x/currencies/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"
	"testing"
)

const (
	queryCurrencyGetIssuePath    = "/custom/currencies/" + currencies.QueryGetIssue
	queryCurrencyGetCurrencyPath = "/custom/currencies/" + currencies.QueryGetCurrency
	queryCurrencyGetDestroyPath  = "/custom/currencies/" + currencies.QueryGetDestroy
	queryCurrencyGetDestroysPath = "/custom/currencies/" + currencies.QueryGetDestroys
)

func Test_CurrencyQueries(t *testing.T) {
	app, server := newTestWbApp()
	defer app.CloseConnections()
	defer server.Stop()

	genCoins, err := sdk.ParseCoins("1000000000000000wings")
	require.NoError(t, err)
	genAccs, _, _, genPrivKeys := CreateGenAccounts(10, genCoins)

	_, err = setGenesis(t, app, genAccs)
	require.NoError(t, err)

	recipientIdx, recipientAddr, recipientPrivKey := uint(0), genAccs[0].Address, genPrivKeys[0]

	checkCurrencyQueryObj := func(obj curTypes.Currency, symbol string, amount int64, decimals int8) {
		require.Equal(t, obj.Symbol, symbol)
		require.Equal(t, obj.Supply.Int64(), amount)
		require.Equal(t, obj.Decimals, decimals)
	}

	checkDestroyQueryObj := func(obj curTypes.Destroy, id int64, symbol string, amount int64, recipientAddr sdk.AccAddress) {
		require.Equal(t, int64(0), obj.ID.Int64())
		require.Equal(t, chainID, obj.ChainID)
		require.Equal(t, symbol, obj.Symbol)
		require.Equal(t, amount, obj.Amount.Int64())
		require.Equal(t, recipientAddr, obj.Spender)
	}

	// issue multiple currencies
	checkCurrencyIssued(t, app, currency1Symbol, amount, 0, "msg1", issue1ID, recipientIdx, genAccs, genPrivKeys)
	checkCurrencyIssued(t, app, currency2Symbol, amount, 0, "msg2", issue2ID, recipientIdx, genAccs, genPrivKeys)
	checkCurrencyIssued(t, app, currency3Symbol, amount, 0, "msg3", issue3ID, recipientIdx, genAccs, genPrivKeys)

	// check getCurrency query
	{
		checkSymbol := func(symbol string) {
			currencyObj := curTypes.Currency{}
			CheckRunQuery(t, app, curTypes.CurrencyReq{Symbol: symbol}, queryCurrencyGetCurrencyPath, &currencyObj)
			checkCurrencyQueryObj(currencyObj, symbol, amount, 0)
		}

		checkSymbol(currency1Symbol)
		checkSymbol(currency2Symbol)
		checkSymbol(currency3Symbol)
	}

	// destroy currency
	checkCurrencyDestroyed(t, app, chainID, currency3Symbol, amount, recipientAddr, recipientPrivKey)
	destroyID := int64(0)

	// check destroys query with pagination
	{
		destroys := curTypes.Destroys{}
		CheckRunQuery(t, app, curTypes.DestroysReq{Page: sdk.NewInt(1), Limit: sdk.NewInt(1)}, queryCurrencyGetDestroysPath, &destroys)

		require.Len(t, destroys, 1)
		checkDestroyQueryObj(destroys[0], destroyID, currency3Symbol, amount, recipientAddr)
	}

	// check single destroy query
	{
		destroy := curTypes.Destroy{}
		CheckRunQuery(t, app, curTypes.DestroyReq{DestroyId: sdk.NewInt(destroyID)}, queryCurrencyGetDestroyPath, &destroy)

		checkDestroyQueryObj(destroy, destroyID, currency3Symbol, amount, recipientAddr)
	}

	// check non-existing currency query
	{
		CheckRunQuerySpecificError(t, app, curTypes.CurrencyReq{Symbol: "non-existing-symbol"}, queryCurrencyGetCurrencyPath, curTypes.ErrNotExistCurrency(""))
	}
}

func Test_CurrencyIssue(t *testing.T) {
	app, server := newTestWbApp()
	defer app.CloseConnections()
	defer server.Stop()

	genCoins, err := sdk.ParseCoins("1000000000000000wings")
	require.NoError(t, err)
	genAccs, _, _, genPrivKeys := CreateGenAccounts(10, genCoins)

	_, err = setGenesis(t, app, genAccs)
	require.NoError(t, err)

	recipientIdx, recipientAddr := uint(0), genAccs[0].Address

	// check currency is issued
	{
		checkCurrencyIssued(t, app, currency1Symbol, amount, 0, "1", issue1ID, recipientIdx, genAccs, genPrivKeys)
		checkIssueExists(t, app, issue1ID, currency1Symbol, amount, recipientAddr)
		checkCurrencyExists(t, app, currency1Symbol, amount, 0)
	}

	// check currency supply increased
	{
		newAmount := int64(200)
		checkCurrencyIssued(t, app, currency1Symbol, newAmount, 0, "2", issue2ID, recipientIdx, genAccs, genPrivKeys)
		checkIssueExists(t, app, issue2ID, currency1Symbol, newAmount, recipientAddr)
		checkCurrencyExists(t, app, currency1Symbol, amount+newAmount, 0)
	}
}

func Test_CurrencyDestroy(t *testing.T) {
	app, server := newTestWbApp()
	defer app.CloseConnections()
	defer server.Stop()

	genCoins, err := sdk.ParseCoins("1000000000000000wings")
	require.NoError(t, err)
	genAccs, _, _, genPrivKeys := CreateGenAccounts(10, genCoins)

	_, err = setGenesis(t, app, genAccs)
	require.NoError(t, err)

	recipientIdx, recipientAddr, recipientPrivKey := uint(0), genAccs[0].Address, genPrivKeys[0]
	curSupply := amount * 2

	// check currency is issued
	{
		checkCurrencyIssued(t, app, currency1Symbol, curSupply, 0, "1", issue1ID, recipientIdx, genAccs, genPrivKeys)
		checkIssueExists(t, app, issue1ID, currency1Symbol, curSupply, recipientAddr)
		checkCurrencyExists(t, app, currency1Symbol, curSupply, 0)
	}

	// check currency supply reduced
	{
		curSupply -= amount
		checkCurrencyDestroyed(t, app, chainID, currency1Symbol, amount, recipientAddr, recipientPrivKey)
		checkCurrencyExists(t, app, currency1Symbol, curSupply, 0)
	}

	// check currency destroyed
	{
		curSupply -= amount
		checkCurrencyDestroyed(t, app, chainID, currency1Symbol, amount, recipientAddr, recipientPrivKey)
		checkCurrencyExists(t, app, currency1Symbol, 0, 0)
	}

	// check currency destroy over the limit
	{
		recipientAcc := GetAccountCheckTx(app, recipientAddr)
		destroyMsg := curMsgs.NewMsgDestroyCurrency(chainID, currency1Symbol, sdk.NewInt(1), recipientAddr, recipientAddr.String())
		tx := genTx([]sdk.Msg{destroyMsg}, []uint64{recipientAcc.GetAccountNumber()}, []uint64{recipientAcc.GetSequence()}, recipientPrivKey)
		CheckDeliverSpecificErrorTx(t, app, tx, sdk.ErrInsufficientCoins(""))
	}
}

func checkCurrencyIssued(t *testing.T, app *WbServiceApp,
	curSymbol string, curAmount int64, curDecimals int8, msgID, issueID string,
	recipientAccIdx uint, accs []*auth.BaseAccount, privKeys []crypto.PrivKey) {

	issueMsg := curMsgs.NewMsgIssueCurrency(curSymbol, sdk.NewInt(curAmount), curDecimals, accs[recipientAccIdx].Address, issueID)
	MSMsgSubmitAndVote(t, app, msgID, issueMsg, recipientAccIdx, accs, privKeys, true)
}

func checkCurrencyDestroyed(t *testing.T, app *WbServiceApp, chainID, curSymbol string, curAmount int64, recipientAddr sdk.AccAddress, recipientPrivKey crypto.PrivKey) {
	recipientAcc := GetAccountCheckTx(app, recipientAddr)
	destroyMsg := curMsgs.NewMsgDestroyCurrency(chainID, curSymbol, sdk.NewInt(curAmount), recipientAcc.GetAddress(), recipientAcc.GetAddress().String())
	tx := genTx([]sdk.Msg{destroyMsg}, []uint64{recipientAcc.GetAccountNumber()}, []uint64{recipientAcc.GetSequence()}, recipientPrivKey)
	CheckDeliverTx(t, app, tx)
}

func checkCurrencyExists(t *testing.T, app *WbServiceApp, symbol string, supply int64, decimals int8) {
	currencyObj := curTypes.Currency{}
	CheckRunQuery(t, app, curTypes.CurrencyReq{Symbol: symbol}, queryCurrencyGetCurrencyPath, &currencyObj)

	require.Equal(t, symbol, currencyObj.Symbol, "symbol")
	require.Equal(t, supply, currencyObj.Supply.Int64(), "amount")
	require.Equal(t, decimals, currencyObj.Decimals, "decimals")
}

func checkIssueExists(t *testing.T, app *WbServiceApp, issueID, symbol string, amount int64, recipientAddr sdk.AccAddress) {
	issue := curTypes.Issue{}
	CheckRunQuery(t, app, curTypes.IssueReq{IssueID: issueID}, queryCurrencyGetIssuePath, &issue)

	require.Equal(t, symbol, issue.Symbol, "symbol")
	require.Equal(t, amount, issue.Amount.Int64(), "amount")
	require.Equal(t, recipientAddr, issue.Recipient)
}
