package app

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"

	"wings-blockchain/x/currencies/msgs"
	"wings-blockchain/x/currencies/types"
	msmsg "wings-blockchain/x/multisig/msgs"
	mstypes "wings-blockchain/x/multisig/types"
	msgspoa "wings-blockchain/x/poa/msgs"
)

const (
	queryIssuePath    = "/custom/currencies/issue"
	queryCurrencyPath = "/custom/currencies/currency"
	queryCallsPath    = "/custom/multisig/calls"
)

var (
	chainID         = ""
	currency1Symbol = "testcoin1"
	currency2Symbol = "testcoin2"
	currency3Symbol = "testcoin3"
	issue1ID        = "issue1"
	issue2ID        = "issue2"
	issue3ID        = "issue3"
	amount          = int64(100)
	ethAddresses    = []string{
		"0x82A978B3f5962A5b0957d9ee9eEf472EE55B42F1",
		"0x7d577a597B2742b498Cb5Cf0C26cDCD726d39E6e",
		"0xDCEceAF3fc5C0a63d195d69b1A90011B7B19650D",
		"0x598443F1880Ef585B21f1d7585Bd0577402861E5",
		"0x13cBB8D99C6C4e0f2728C7d72606e78A29C4E224",
		"0x77dB2BEBBA79Db42a978F896968f4afCE746ea1F",
		"0x24143873e0E0815fdCBcfFDbe09C979CbF9Ad013",
		"0x10A1c1CB95c92EC31D3f22C66Eef1d9f3F258c6B",
		"0xe0FC04FA2d34a66B779fd5CEe748268032a146c0",
	}
)

func issueCurrencyCheck(t *testing.T, app *WbServiceApp, msgID string, msg msgs.MsgIssueCurrency, recipient sdk.AccAddress,
	genAccs []*auth.BaseAccount, addrs []sdk.AccAddress, privKeys []crypto.PrivKey) {

	// Submit message
	submitMsg := msmsg.NewMsgSubmitCall(msg, msgID, recipient)
	{
		acc := GetAccount(app, genAccs[0].Address)
		tx := genTx([]sdk.Msg{submitMsg}, []uint64{acc.GetAccountNumber()}, []uint64{acc.GetSequence()}, privKeys[0])
		CheckDeliverTx(t, app, tx)
	}
	calls := mstypes.CallsResp{}
	CheckRunQuery(t, app, nil, queryCallsPath, &calls)
	require.Equal(t, 1, len(calls[0].Votes))

	// Vote, vote, vote...
	confirmMsg := msmsg.MsgConfirmCall{MsgId: calls[0].Call.MsgID}
	for i := 1; i < len(genAccs)/2; i++ {
		{
			confirmMsg.Sender = addrs[i]
			acc := GetAccount(app, genAccs[i].Address)
			tx := genTx([]sdk.Msg{confirmMsg}, []uint64{acc.GetAccountNumber()}, []uint64{acc.GetSequence()}, privKeys[i])
			CheckDeliverTx(t, app, tx)
		}
		CheckRunQuery(t, app, nil, queryCallsPath, &calls)
		require.Equal(t, i+1, len(calls[0].Votes))
	}

	confirmMsg.Sender = addrs[len(addrs)-1]
	{
		acc := GetAccount(app, genAccs[len(addrs)-1].Address)
		tx := genTx([]sdk.Msg{confirmMsg}, []uint64{acc.GetAccountNumber()}, []uint64{acc.GetSequence()}, privKeys[len(addrs)-1])
		CheckDeliverTx(t, app, tx)
	}
	CheckRunQuery(t, app, nil, queryCallsPath, &calls)
	require.Equal(t, 0, len(calls))
}

func destroyCurrency(t *testing.T, app *WbServiceApp, msg msgs.MsgDestroyCurrency, genAccs []*auth.BaseAccount, addrs []sdk.AccAddress, privKeys []crypto.PrivKey) {
	acc := GetAccount(app, genAccs[0].Address)
	tx := genTx([]sdk.Msg{msg}, []uint64{acc.GetAccountNumber()}, []uint64{acc.GetSequence()}, privKeys[0])
	CheckDeliverTx(t, app, tx)
}

func checkCurrencyExists(t *testing.T, app *WbServiceApp, symbol string, amount int64, decimals int8) {
	var currency types.Currency
	CheckRunQuery(t, app, types.CurrencyReq{Symbol: currency1Symbol}, queryCurrencyPath, &currency)
	require.Equal(t, currency1Symbol, currency.Symbol)
	require.Equal(t, amount, currency.Supply.Int64())
	require.Equal(t, decimals, currency.Decimals)
}

func checkIssueExists(t *testing.T, app *WbServiceApp, issueID string, recipient sdk.AccAddress, amount int64) {
	var issue types.Issue
	CheckRunQuery(t, app, types.IssueReq{IssueID: issueID}, queryIssuePath, &issue)
	require.Equal(t, currency1Symbol, issue.Symbol)
	require.Equal(t, amount, issue.Amount.Int64())
	require.Equal(t, recipient, issue.Recipient)
}

func Test_IssueCurrency(t *testing.T) {
	t.Parallel()

	// preparing test environment
	app := newTestWbApp()
	genCoins, err := sdk.ParseCoins("1000000000000000wings")
	require.NoError(t, err)
	// Create a bunch (ie 10) of pre-funded accounts to use for tests
	genAccs, addrs, _, privKeys := CreateGenAccounts(10, genCoins)
	_, err = setGenesis(t, app, genAccs)
	require.NoError(t, err)

	// issue currency
	recipient := addrs[0]
	issueMsg := msgs.NewMsgIssueCurrency(currency1Symbol, sdk.NewInt(amount), 0, recipient, issue1ID)
	issueCurrencyCheck(t, app, "1", issueMsg, recipient, genAccs, addrs, privKeys)
	// checking that the currency is issued
	checkCurrencyExists(t, app, currency1Symbol, amount, 0)
	// check issue is exists
	checkIssueExists(t, app, issue1ID, recipient, amount)
}

func Test_IssueCurrencyTwice(t *testing.T) {
	t.Parallel()

	// preparing test environment
	app := newTestWbApp()
	genCoins, err := sdk.ParseCoins("1000000000000000wings")
	require.NoError(t, err)
	// Create a bunch (ie 10) of pre-funded accounts to use for tests
	genAccs, addrs, _, privKeys := CreateGenAccounts(10, genCoins)
	_, err = setGenesis(t, app, genAccs)
	require.NoError(t, err)

	// issue currency
	recipient := addrs[0]
	issueMsg := msgs.NewMsgIssueCurrency(currency1Symbol, sdk.NewInt(amount), 0, recipient, issue1ID)
	issueCurrencyCheck(t, app, "1", issueMsg, recipient, genAccs, addrs, privKeys)
	checkIssueExists(t, app, issue1ID, recipient, amount)
	newAmount := int64(200)
	issueMsg.IssueID = issue2ID
	issueMsg.Amount = sdk.NewInt(newAmount)
	issueCurrencyCheck(t, app, "2", issueMsg, recipient, genAccs, addrs, privKeys)
	// checking that the currency is issued
	checkCurrencyExists(t, app, currency1Symbol, amount+newAmount, 0)
	// check issue is exists
	checkIssueExists(t, app, issue2ID, recipient, newAmount)
}

func Test_DestroyCurrency(t *testing.T) {
	t.Parallel()

	app := newTestWbApp()
	genCoins, err := sdk.ParseCoins("1000000000000000wings")
	require.NoError(t, err)

	// Create a bunch (ie 10) of pre-funded accounts to use for tests
	genAccs, addrs, _, privKeys := CreateGenAccounts(10, genCoins)
	_, err = setGenesis(t, app, genAccs)
	require.NoError(t, err)

	recipient := addrs[0]
	issueMsg := msgs.NewMsgIssueCurrency(currency1Symbol, sdk.NewInt(amount), 0, recipient, issue1ID)
	issueCurrencyCheck(t, app, "1", issueMsg, recipient, genAccs, addrs, privKeys)
	// checking that the currency is issued
	checkCurrencyExists(t, app, currency1Symbol, amount, 0)
	// check issue is exists
	checkIssueExists(t, app, issue1ID, recipient, amount)
	destroyMsg := msgs.NewMsgDestroyCurrency(chainID, currency1Symbol, sdk.NewInt(amount), addrs[0], addrs[0].String())
	destroyCurrency(t, app, destroyMsg, genAccs, addrs, privKeys)
	checkCurrencyExists(t, app, currency1Symbol, 0, 0)
}

func Test_Queryes(t *testing.T) {
	t.Parallel()

	app := newTestWbApp()
	genCoins, err := sdk.ParseCoins("1000000000000000wings")
	require.NoError(t, err)

	// Create a bunch (ie 10) of pre-funded accounts to use for tests
	genAccs, addrs, _, privKeys := CreateGenAccounts(10, genCoins)
	_, err = setGenesis(t, app, genAccs)
	require.NoError(t, err)

	recipient := addrs[0]
	issue1Msg := msgs.NewMsgIssueCurrency(currency1Symbol, sdk.NewInt(amount), 0, recipient, issue1ID)
	issueCurrencyCheck(t, app, "msg1", issue1Msg, recipient, genAccs, addrs, privKeys)

	issue2Msg := msgs.NewMsgIssueCurrency(currency2Symbol, sdk.NewInt(amount), 0, recipient, issue2ID)
	issueCurrencyCheck(t, app, "msg2", issue2Msg, recipient, genAccs, addrs, privKeys)

	issue3Msg := msgs.NewMsgIssueCurrency(currency3Symbol, sdk.NewInt(amount), 0, recipient, issue3ID)
	issueCurrencyCheck(t, app, "msg3", issue3Msg, recipient, genAccs, addrs, privKeys)

	destroyMsg := msgs.NewMsgDestroyCurrency(chainID, currency3Symbol, sdk.NewInt(amount), addrs[0], addrs[0].String())
	destroyCurrency(t, app, destroyMsg, genAccs, addrs, privKeys)
}

func Test_POAHandlerIsMultisigOnly(t *testing.T) {
	t.Parallel()

	app := newTestWbApp()
	genCoins, err := sdk.ParseCoins("1000000000000000wings")
	require.NoError(t, err)
	genValidators, _, _, privKeys := CreateGenAccounts(7, genCoins)
	_, err = setGenesis(t, app, genValidators)
	require.NoError(t, err)
	validatorsToAdd, _, _, _ := CreateGenAccounts(1, genCoins)
	addValidatorMsg := msgspoa.NewMsgAddValidator(validatorsToAdd[0].Address, ethAddresses[0], genValidators[0].Address)
	acc := GetAccount(app, genValidators[0].Address)
	tx := genTx([]sdk.Msg{addValidatorMsg}, []uint64{acc.GetAccountNumber()}, []uint64{acc.GetSequence()}, privKeys[0])
	CheckDeliverErrorTx(t, app, tx)
}

func CheckAddValidators(t *testing.T, app *WbServiceApp, genAccs []*auth.BaseAccount, newValidators []*auth.BaseAccount, privKeys []crypto.PrivKey) {
	sender := genAccs[0].Address
	for _, v := range newValidators {
		// Submit message
		addValidatorMsg := msgspoa.NewMsgAddValidator(v.Address, ethAddresses[0], sender)
		acc := GetAccount(app, sender)
		msgID := fmt.Sprintf("addValidator:%s", v.Address)
		submitMsg := msmsg.NewMsgSubmitCall(addValidatorMsg, msgID, sender)
		tx := genTx([]sdk.Msg{submitMsg}, []uint64{acc.GetAccountNumber()}, []uint64{acc.GetSequence()}, privKeys[0])
		CheckDeliverTx(t, app, tx)

		calls := mstypes.CallsResp{}
		CheckRunQuery(t, app, nil, queryCallsPath, &calls)
		require.Equal(t, 1, len(calls[0].Votes))
		confirmMsg := msmsg.MsgConfirmCall{MsgId: calls[0].Call.MsgID}
		validatorsAmount := app.poaKeeper.GetValidatorAmount(GetContext(app, true))
		for idx, vv := range genAccs[1 : validatorsAmount/2+1] {
			acc := GetAccount(app, vv.Address)
			confirmMsg.Sender = vv.Address
			tx := genTx([]sdk.Msg{confirmMsg}, []uint64{acc.GetAccountNumber()}, []uint64{acc.GetSequence()}, privKeys[idx+1])
			CheckDeliverTx(t, app, tx)
		}
	}
}

func Test_POAValidatorsAdd(t *testing.T) {
	t.Parallel()

	app := newTestWbApp()
	genCoins, err := sdk.ParseCoins("1000000000000000wings")
	require.NoError(t, err)
	genValidators, _, _, privKeys := CreateGenAccounts(7, genCoins)
	_, err = setGenesis(t, app, genValidators)
	require.NoError(t, err)

	validatorsToAdd, _, _, _ := CreateGenAccounts(4, genCoins)
	CheckAddValidators(t, app, genValidators, validatorsToAdd, privKeys)
	var added int
	validators := app.poaKeeper.GetValidators(GetContext(app, true))
Loop:
	for _, v := range validatorsToAdd {
		for _, vv := range validators {
			if v.Address.String() == vv.Address.String() {
				added++
				continue Loop
			}
		}
	}
	require.Equal(t, added, len(validatorsToAdd))
}
