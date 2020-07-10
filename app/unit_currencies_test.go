// +build unit

package app

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/dfinance/dnode/x/ccstorage"
	"github.com/dfinance/dnode/x/currencies"
	"github.com/dfinance/dnode/x/multisig"
)

// Checks that currencies module supports only multisig calls for issue msg (using MSRouter).
func TestCurrenciesApp_MultisigHandler(t *testing.T) {
	t.Parallel()

	app, appStop := NewTestDnAppMockVM()
	defer appStop()

	genValidators, _, _, genPrivKeys := CreateGenAccounts(7, GenDefCoins(t))
	CheckSetGenesisMockVM(t, app, genValidators)

	{
		senderAcc, senderPrivKey := GetAccountCheckTx(app, genValidators[0].Address), genPrivKeys[0]
		issueMsg := currencies.NewMsgIssueCurrency(issue1ID, coin1, senderAcc.GetAddress())
		tx := GenTx([]sdk.Msg{issueMsg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
		CheckDeliverSpecificErrorTx(t, app, tx, sdkErrors.ErrUnauthorized)
	}
}

// Test currencies module queries.
func TestCurrenciesApp_Queries(t *testing.T) {
	t.Parallel()

	app, appStop := NewTestDnAppMockVM()
	defer appStop()

	genAccs, _, _, genPrivKeys := CreateGenAccounts(10, GenDefCoins(t))
	CheckSetGenesisMockVM(t, app, genAccs)

	recipientIdx, recipientAddr, recipientPrivKey := uint(0), genAccs[0].Address, genPrivKeys[0]

	checkWithdrawQueryObj := func(obj currencies.Withdraw, id uint64, coin sdk.Coin, spenderAddr sdk.AccAddress) {
		require.Equal(t, id, obj.ID.UInt64())
		require.Equal(t, coin.Denom, obj.Coin.Denom)
		require.True(t, coin.Amount.Equal(obj.Coin.Amount))
		require.Equal(t, spenderAddr, obj.Spender)
		require.Equal(t, chainID, obj.PegZoneChainID)
	}

	// issue multiple currencies
	CreateCurrency(t, app, currency1Denom, 0)
	CreateCurrency(t, app, currency2Denom, 0)
	CreateCurrency(t, app, currency3Denom, 0)
	IssueCurrency(t, app, coin1, "msg1", issue1ID, recipientIdx, genAccs, genPrivKeys, true)
	IssueCurrency(t, app, coin2, "msg2", issue2ID, recipientIdx, genAccs, genPrivKeys, true)
	IssueCurrency(t, app, coin3, "msg3", issue3ID, recipientIdx, genAccs, genPrivKeys, true)

	// check getCurrency query
	{
		CheckCurrencyExists(t, app, currency1Denom, amount, 0)
		CheckCurrencyExists(t, app, currency2Denom, amount, 0)
		CheckCurrencyExists(t, app, currency3Denom, amount, 0)
	}

	// check getIssue query
	{
		CheckIssueExists(t, app, issue1ID, coin1, recipientAddr)
		CheckIssueExists(t, app, issue2ID, coin2, recipientAddr)
		CheckIssueExists(t, app, issue3ID, coin3, recipientAddr)
	}

	// withdraw currencies
	withdrawAmount := amount.QuoRaw(3)
	withdrawCoin := sdk.NewCoin(currency3Denom, withdrawAmount)
	WithdrawCurrency(t, app, chainID, withdrawCoin, recipientAddr, recipientPrivKey, true)
	WithdrawCurrency(t, app, chainID, withdrawCoin, recipientAddr, recipientPrivKey, true)
	WithdrawCurrency(t, app, chainID, withdrawCoin, recipientAddr, recipientPrivKey, true)

	// check getWithdraws query with pagination
	{
		// page 1
		{
			withdraws := currencies.Withdraws{}
			reqParams := currencies.WithdrawsReq{Page: sdk.NewUint(1), Limit: sdk.NewUint(2)}
			CheckRunQuery(t, app, reqParams, queryCurrencyWithdrawsPath, &withdraws)

			require.Len(t, withdraws, 2)
			checkWithdrawQueryObj(withdraws[0], 0, withdrawCoin, recipientAddr)
			checkWithdrawQueryObj(withdraws[1], 1, withdrawCoin, recipientAddr)
		}

		// page 2
		{
			withdraws := currencies.Withdraws{}
			reqParams := currencies.WithdrawsReq{Page: sdk.NewUint(2), Limit: sdk.NewUint(2)}
			CheckRunQuery(t, app, reqParams, queryCurrencyWithdrawsPath, &withdraws)

			require.Len(t, withdraws, 1)
			checkWithdrawQueryObj(withdraws[0], 2, withdrawCoin, recipientAddr)
		}
	}

	// check getWithdraw query
	{
		CheckWithdrawExists(t, app, 0, withdrawCoin, recipientAddr, recipientAddr.String())
		CheckWithdrawExists(t, app, 1, withdrawCoin, recipientAddr, recipientAddr.String())
		CheckWithdrawExists(t, app, 2, withdrawCoin, recipientAddr, recipientAddr.String())
	}
}

// Test currency issue logic with failure scenarios.
func TestCurrenciesApp_Issue(t *testing.T) {
	t.Parallel()

	app, appStop := NewTestDnAppMockVM()
	defer appStop()

	genAccs, _, _, genPrivKeys := CreateGenAccounts(10, GenDefCoins(t))
	CheckSetGenesisMockVM(t, app, genAccs)

	recipientIdx, recipientAddr := uint(0), genAccs[0].Address
	curAmount, curDecimals, denom := amount, uint8(0), currency1Denom

	CreateCurrency(t, app, denom, curDecimals)

	// ok: currency is issued
	{
		msgId, issueId := "1", "issue1"
		coin := sdk.NewCoin(denom, curAmount)

		IssueCurrency(t, app, coin, msgId, issueId, recipientIdx, genAccs, genPrivKeys, true)
		CheckIssueExists(t, app, issueId, coin, recipientAddr)
		CheckCurrencyExists(t, app, denom, curAmount, curDecimals)
		CheckRecipientCoins(t, app, recipientAddr, denom, curAmount)
	}

	// ok currency supply increased
	{
		msgId, issueId := "2", "issue2"
		newAmount := sdk.NewInt(200)
		coin := sdk.NewCoin(denom, newAmount)
		curAmount = curAmount.Add(newAmount)

		IssueCurrency(t, app, coin, msgId, issueId, recipientIdx, genAccs, genPrivKeys, true)
		CheckIssueExists(t, app, issueId, coin, recipientAddr)
		CheckCurrencyExists(t, app, denom, curAmount, curDecimals)
		CheckRecipientCoins(t, app, recipientAddr, denom, curAmount)
	}

	// fail: currency issue with the same issueID
	{
		msgId, issueId := "non-existing-msgID", "issue1"
		coin := sdk.NewCoin(denom, amount)

		res, err := IssueCurrency(t, app, coin, msgId, issueId, recipientIdx, genAccs, genPrivKeys, false)
		CheckResultError(t, currencies.ErrWrongIssueID, res, err)
	}

	// fail: currency issue with already existing uniqueMsgID
	{
		msgId, issueId := "1", "non-existing-issue"
		coin := sdk.NewCoin(denom, amount)

		res, err := IssueCurrency(t, app, coin, msgId, issueId, recipientIdx, genAccs, genPrivKeys, false)
		CheckResultError(t, multisig.ErrWrongCallUniqueId, res, err)
	}
}

// Test maximum bank supply level (DVM has u128 limit).
func TestCurrenciesApp_IssueHugeAmount(t *testing.T) {
	t.Parallel()

	app, appStop := NewTestDnAppMockVM()
	defer appStop()

	genAccs, _, _, genPrivKeys := CreateGenAccounts(10, GenDefCoins(t))
	CheckSetGenesisMockVM(t, app, genAccs)

	recipientIdx, recipientAddr := uint(0), genAccs[0].Address

	// check huge amount currency issue (max value for u128)
	{
		msgId, issueId, denom := "1", "issue1", currency1Denom

		hugeAmount, ok := sdk.NewIntFromString("100000000000000000000000000000000000000")
		require.True(t, ok)
		coin := sdk.NewCoin(denom, hugeAmount)

		CreateCurrency(t, app, denom, 0)
		IssueCurrency(t, app, coin, msgId, issueId, recipientIdx, genAccs, genPrivKeys, true)
		CheckIssueExists(t, app, issueId, coin, recipientAddr)
		CheckCurrencyExists(t, app, denom, hugeAmount, 0)
		CheckRecipientCoins(t, app, recipientAddr, denom, hugeAmount)
	}

	// check huge amount currency issue (that worked before u128)
	{
		msgId, issueId, denom := "2", "issue2", currency2Denom

		hugeAmount, ok := sdk.NewIntFromString("1000000000000000000000000000000000000000000000")
		require.True(t, ok)
		coin := sdk.NewCoin(denom, hugeAmount)

		CreateCurrency(t, app, denom, 0)
		IssueCurrency(t, app, coin, msgId, issueId, recipientIdx, genAccs, genPrivKeys, true)
		CheckIssueExists(t, app, issueId, coin, recipientAddr)
		CheckCurrencyExists(t, app, denom, hugeAmount, 0)

		require.Panics(t, func() {
			app.bankKeeper.GetCoins(GetContext(app, true), recipientAddr)
		})
	}
}

// Test issue/withdraw currency with decimals.
func TestCurrenciesApp_Decimals(t *testing.T) {
	t.Parallel()

	app, appStop := NewTestDnAppMockVM()
	defer appStop()

	genAccs, _, _, genPrivKeys := CreateGenAccounts(10, GenDefCoins(t))
	CheckSetGenesisMockVM(t, app, genAccs)

	recipientIdx, recipientAddr, recipientPrivKey := uint(0), genAccs[0].Address, genPrivKeys[0]
	curAmount, curDecimals, denom := sdk.OneInt(), uint8(1), currency1Denom

	CreateCurrency(t, app, denom, curDecimals)

	// issue currency amount with decimals
	{
		msgId, issueId := "1", "issue1"
		coin := sdk.NewCoin(denom, curAmount)

		IssueCurrency(t, app, coin, msgId, issueId, recipientIdx, genAccs, genPrivKeys, true)
		CheckIssueExists(t, app, issueId, coin, recipientAddr)
		CheckCurrencyExists(t, app, denom, curAmount, curDecimals)
		CheckRecipientCoins(t, app, recipientAddr, denom, curAmount)
	}

	// increase currency amount with decimals
	{
		msgId, issueId := "2", "issue2"

		newAmount := sdk.OneInt()
		coin := sdk.NewCoin(denom, newAmount)
		curAmount = curAmount.Add(newAmount)

		IssueCurrency(t, app, coin, msgId, issueId, recipientIdx, genAccs, genPrivKeys, true)
		CheckIssueExists(t, app, issueId, coin, recipientAddr)
		CheckCurrencyExists(t, app, denom, curAmount, curDecimals)
		CheckRecipientCoins(t, app, recipientAddr, denom, curAmount)
	}

	// decrease currency amount with decimals
	{
		newAmount := sdk.OneInt()
		coin := sdk.NewCoin(denom, newAmount)
		curAmount = curAmount.Sub(newAmount)

		WithdrawCurrency(t, app, chainID, coin, recipientAddr, recipientPrivKey, true)
		CheckCurrencyExists(t, app, denom, curAmount, curDecimals)
		CheckRecipientCoins(t, app, recipientAddr, denom, curAmount)
	}
}

// Test withdraw currency with fail scenarios.
func TestCurrenciesApp_Withdraw(t *testing.T) {
	t.Parallel()

	app, appStop := NewTestDnAppMockVM()
	defer appStop()

	genAccs, _, _, genPrivKeys := CreateGenAccounts(10, GenDefCoins(t))
	CheckSetGenesisMockVM(t, app, genAccs)

	recipientIdx, recipientAddr, recipientPrivKey := uint(0), genAccs[0].Address, genPrivKeys[0]
	curSupply, denom := amount.Mul(sdk.NewInt(2)), currency1Denom

	CreateCurrency(t, app, denom, 0)

	// issue currency
	{
		coin := sdk.NewCoin(denom, curSupply)
		IssueCurrency(t, app, coin, "1", issue1ID, recipientIdx, genAccs, genPrivKeys, true)
		CheckIssueExists(t, app, issue1ID, coin, recipientAddr)
		CheckCurrencyExists(t, app, denom, curSupply, 0)
		CheckRecipientCoins(t, app, recipientAddr, denom, curSupply)
	}

	// ok: withdraw currency
	{
		coin := sdk.NewCoin(denom, amount)
		curSupply = curSupply.Sub(amount)
		WithdrawCurrency(t, app, chainID, coin, recipientAddr, recipientPrivKey, true)
		CheckWithdrawExists(t, app, 0, coin, recipientAddr, recipientAddr.String())
		CheckCurrencyExists(t, app, denom, curSupply, 0)
		CheckRecipientCoins(t, app, recipientAddr, denom, curSupply)
	}

	// ok: withdraw currency (currency supply is 0)
	{
		coin := sdk.NewCoin(denom, amount)
		curSupply = curSupply.Sub(amount)
		require.True(t, curSupply.IsZero())

		WithdrawCurrency(t, app, chainID, coin, recipientAddr, recipientPrivKey, true)
		CheckWithdrawExists(t, app, 1, coin, recipientAddr, recipientAddr.String())
		CheckCurrencyExists(t, app, denom, curSupply, 0)
		CheckRecipientCoins(t, app, recipientAddr, denom, curSupply)
	}

	// fail: currency withdraw over the limit
	{
		coin := sdk.NewCoin(denom, sdk.OneInt())
		res, err := WithdrawCurrency(t, app, chainID, coin, recipientAddr, recipientPrivKey, false)
		CheckResultError(t, sdkErrors.ErrInsufficientFunds, res, err)
	}

	// fail: currency withdraw with denom account doesn't have
	{
		wrongDenom := currency2Denom
		coin := sdk.NewCoin(wrongDenom, sdk.OneInt())

		res, err := WithdrawCurrency(t, app, chainID, coin, recipientAddr, recipientPrivKey, false)
		CheckResultError(t, ccstorage.ErrWrongDenom, res, err)
	}
}

// Test issues and destroys currency and verifies that supply (via supply module) stays up-to-date.
func TestCurrenciesApp_Supply(t *testing.T) {
	t.Parallel()

	app, appStop := NewTestDnAppMockVM()
	defer appStop()

	genAccs, _, _, genPrivKeys := CreateGenAccounts(10, GenDefCoins(t))
	CheckSetGenesisMockVM(t, app, genAccs)

	checkSupplies := func(testID string) {
		ctx := GetContext(app, true)
		supplies := GetAllSupplies(t, app, ctx)
		if err := supplies.AreEqual(); err != nil {
			t.Logf(">> TestCase: %s", testID)
			t.Log(supplies.String())
			require.NoError(t, err)
		}
	}

	// initial check
	{
		checkSupplies("initial")
	}

	// issue 50.0 dfi to account1
	{
		amount, _ := sdk.NewIntFromString("50000000000000000000")
		coin := sdk.NewCoin("dfi", amount)
		IssueCurrency(t, app, coin, "1", issue1ID, uint(0), genAccs, genPrivKeys, true)

		checkSupplies("50.0 dfi issued to acc #1")
	}

	// issue 5.0 btc to account2
	{
		amount, _ := sdk.NewIntFromString("500000000")
		coin := sdk.NewCoin("btc", amount)
		IssueCurrency(t, app, coin, "2", issue2ID, uint(1), genAccs, genPrivKeys, true)

		checkSupplies("5.0 btc issued to acc #2")
	}

	// withdraw 2.5 btc from account2
	{
		recipientAddr, recipientPrivKey := genAccs[1].Address, genPrivKeys[1]
		amount, _ := sdk.NewIntFromString("250000000")
		coin := sdk.NewCoin("btc", amount)
		WithdrawCurrency(t, app, chainID, coin, recipientAddr, recipientPrivKey, true)

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
