package app

import (
	"fmt"
	poaTypes "github.com/WingsDao/wings-blockchain/x/poa/types"
	"strconv"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"

	"github.com/WingsDao/wings-blockchain/x/currencies"
	"github.com/WingsDao/wings-blockchain/x/currencies/msgs"
	"github.com/WingsDao/wings-blockchain/x/currencies/types"
	msmsg "github.com/WingsDao/wings-blockchain/x/multisig/msgs"
	mstypes "github.com/WingsDao/wings-blockchain/x/multisig/types"
	msgspoa "github.com/WingsDao/wings-blockchain/x/poa/msgs"
)

const (
	queryGetIssuePath    = "/custom/currencies/" + currencies.QueryGetIssue
	queryGetCurrencyPath = "/custom/currencies/" + currencies.QueryGetCurrency
	queryGetDestroyPath  = "/custom/currencies/" + currencies.QueryGetDestroy
	queryGetDestroysPath = "/custom/currencies/" + currencies.QueryGetDestroys
	queryGetCallPath     = "/custom/multisig/call"
	queryGetCallsPath    = "/custom/multisig/calls"
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

func Test_IssueCurrency(t *testing.T) {
	// preparing test environment
	app, server := newTestWbApp()
	defer app.CloseConnections()
	defer server.Stop()

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
	app, server := newTestWbApp()
	defer app.CloseConnections()
	defer server.Stop()

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
	app, server := newTestWbApp()
	defer app.CloseConnections()
	defer server.Stop()

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
	destroyCurrency(t, app, destroyMsg, genAccs, privKeys)
	checkCurrencyExists(t, app, currency1Symbol, 0, 0)
}

func Test_DestroyCurrencyOverLimit(t *testing.T) {
	app, server := newTestWbApp()
	defer app.CloseConnections()
	defer server.Stop()

	genCoins, err := sdk.ParseCoins("1000000000000000wings")
	require.NoError(t, err)

	// Create a bunch (ie 10) of pre-funded accounts to use for tests
	genAccs, addrs, _, privKeys := CreateGenAccounts(10, genCoins)
	_, err = setGenesis(t, app, genAccs)
	require.NoError(t, err)

	recipientAddr, recepientPrivKey := addrs[0], privKeys[0]
	issueMsg := msgs.NewMsgIssueCurrency(currency1Symbol, sdk.NewInt(amount), 0, recipientAddr, issue1ID)
	issueCurrencyCheck(t, app, "1", issueMsg, recipientAddr, genAccs, addrs, privKeys)
	// checking that the currency is issued
	checkCurrencyExists(t, app, currency1Symbol, amount, 0)
	// check issue is exists
	checkIssueExists(t, app, issue1ID, recipientAddr, amount)

	// reduce the currency over the limit
	destroyMsg := msgs.NewMsgDestroyCurrency(chainID, currency1Symbol, sdk.NewInt(amount+1), recipientAddr, recipientAddr.String())
	acc := GetAccountCheckTx(app, recipientAddr)
	tx := genTx([]sdk.Msg{destroyMsg}, []uint64{acc.GetAccountNumber()}, []uint64{acc.GetSequence()}, recepientPrivKey)
	CheckDeliverSpecificErrorTx(t, app, tx, sdk.ErrInsufficientCoins(""))
}

func Test_Queryes(t *testing.T) {
	app, server := newTestWbApp()
	defer app.CloseConnections()
	defer server.Stop()

	genCoins, err := sdk.ParseCoins("1000000000000000wings")
	require.NoError(t, err)
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

	var currency types.Currency
	CheckRunQuery(t, app, types.CurrencyReq{Symbol: currency1Symbol}, queryGetCurrencyPath, &currency)
	require.Equal(t, currency.Symbol, currency1Symbol)
	CheckRunQuery(t, app, types.CurrencyReq{Symbol: currency2Symbol}, queryGetCurrencyPath, &currency)
	require.Equal(t, currency.Symbol, currency2Symbol)
	CheckRunQuery(t, app, types.CurrencyReq{Symbol: currency3Symbol}, queryGetCurrencyPath, &currency)
	require.Equal(t, currency.Symbol, currency3Symbol)

	destroyMsg := msgs.NewMsgDestroyCurrency(chainID, currency3Symbol, sdk.NewInt(amount), addrs[0], addrs[0].String())
	destroyCurrency(t, app, destroyMsg, genAccs, privKeys)

	var destroys types.Destroys
	CheckRunQuery(t, app, types.DestroysReq{Page: sdk.NewInt(1), Limit: sdk.NewInt(1)}, queryGetDestroysPath, &destroys)
	require.Equal(t, int(1), len(destroys))
	require.Equal(t, sdk.NewInt(0).Int64(), destroys[0].ID.Int64())
	require.Equal(t, currency3Symbol, destroys[0].Symbol)

	var destroy types.Destroy
	CheckRunQuery(t, app, types.DestroyReq{DestroyId: destroys[0].ID}, queryGetDestroyPath, &destroy)
	require.Equal(t, sdk.NewInt(0).Int64(), destroy.ID.Int64())
	require.Equal(t, currency3Symbol, destroy.Symbol)

}

func Test_POAHandlerIsMultisigOnly(t *testing.T) {
	app, server := newTestWbApp()
	defer app.CloseConnections()
	defer server.Stop()

	genCoins, err := sdk.ParseCoins("1000000000000000wings")
	require.NoError(t, err)
	genValidators, _, _, privKeys := CreateGenAccounts(7, genCoins)
	_, err = setGenesis(t, app, genValidators)
	require.NoError(t, err)
	validatorsToAdd, _, _, _ := CreateGenAccounts(1, genCoins)
	addValidatorMsg := msgspoa.NewMsgAddValidator(validatorsToAdd[0].Address, ethAddresses[0], genValidators[0].Address)
	acc := GetAccountCheckTx(app, genValidators[0].Address)
	tx := genTx([]sdk.Msg{addValidatorMsg}, []uint64{acc.GetAccountNumber()}, []uint64{acc.GetSequence()}, privKeys[0])
	CheckDeliverErrorTx(t, app, tx)
}

func Test_POAValidatorsAdd(t *testing.T) {
	app, server := newTestWbApp()
	defer app.CloseConnections()
	defer server.Stop()

	genCoins, err := sdk.ParseCoins("1000000000000000wings")
	require.NoError(t, err)
	genValidators, _, _, privKeys := CreateGenAccounts(7, genCoins)
	_, err = setGenesis(t, app, genValidators)
	require.NoError(t, err)

	// add new validators
	validatorsToAdd, _, _, _ := CreateGenAccounts(4, genCoins)
	checkAddValidators(t, app, genValidators, validatorsToAdd, privKeys)
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
	require.Equal(t, len(validatorsToAdd)+len(genValidators), len(validators))

	// add already existing validator
	res := addValidators(t, app, genValidators, []*auth.BaseAccount{validatorsToAdd[0]}, privKeys, false)
	CheckResultError(t, poaTypes.ErrValidatorExists(""), res)
}

func Test_POAValidatorsRemove(t *testing.T) {
	app, server := newTestWbApp()
	defer app.CloseConnections()
	defer server.Stop()

	genCoins, err := sdk.ParseCoins("1000000000000000wings")
	require.NoError(t, err)
	genValidators, _, _, privKeys := CreateGenAccounts(7, genCoins)
	_, err = setGenesis(t, app, genValidators)
	require.NoError(t, err)

	validatorsToRemove, _, _, _ := CreateGenAccounts(4, genCoins)
	checkAddValidators(t, app, genValidators, validatorsToRemove, privKeys)
	require.Equal(t, len(genValidators)+len(validatorsToRemove), int(app.poaKeeper.GetValidatorAmount(GetContext(app, true))))

	var added int
	validators := app.poaKeeper.GetValidators(GetContext(app, true))
Loop:
	for _, v := range validatorsToRemove {
		for _, vv := range validators {
			if v.Address.String() == vv.Address.String() {
				added++
				continue Loop
			}
		}
	}
	require.Equal(t, added, len(validatorsToRemove))

	checkRemoveValidators(t, app, genValidators, validatorsToRemove, privKeys)
	require.Equal(t, len(genValidators), int(app.poaKeeper.GetValidatorAmount(GetContext(app, true))))

	// check requested (validatorsToRemove) validator were removed
	existingValidators := append([]*auth.BaseAccount(nil), genValidators...)
	for _, v := range app.poaKeeper.GetValidators(GetContext(app, true)) {
		for ii, vv := range existingValidators {
			if v.Address.Equals(vv.Address) {
				existingValidators = append(existingValidators[:ii], existingValidators[ii+1:]...)
				break
			}
		}
	}
	require.Equal(t, len(existingValidators), 0)

	// remove non-existing validator
	res := removeValidators(t, app, genValidators, []*auth.BaseAccount{validatorsToRemove[0]}, privKeys, false)
	CheckResultError(t, poaTypes.ErrValidatorDoesntExists(""), res)
}

func Test_POAValidatorsReplace(t *testing.T) {
	app, server := newTestWbApp()
	defer app.CloseConnections()
	defer server.Stop()

	genCoins, err := sdk.ParseCoins("1000000000000000wings")
	require.NoError(t, err)
	genValidators, _, _, privKeys := CreateGenAccounts(7, genCoins)
	_, err = setGenesis(t, app, genValidators)
	require.NoError(t, err)

	validatorsToReplace, _, _, _ := CreateGenAccounts(1, genCoins)
	app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{ChainID: chainID, Height: app.LastBlockHeight() + 1}})
	startLen := len(genValidators)
	for idx, acc := range validatorsToReplace {
		acc.AccountNumber = uint64(startLen + idx)
		app.accountKeeper.SetAccount(GetContext(app, false), acc)
	}
	app.EndBlock(abci.RequestEndBlock{})
	app.Commit()

	checkReplaceValidators(t, app, genValidators, validatorsToReplace[0], privKeys)
	// check "new" validator was added ("old" replaced)
	replaced := app.poaKeeper.GetValidator(GetContext(app, true), validatorsToReplace[0].Address)
	require.Equal(t, validatorsToReplace[0].Address.String(), replaced.Address.String())
	require.Equal(t, len(genValidators), int(app.poaKeeper.GetValidatorAmount(GetContext(app, true))))
	// check "old" validator doesn't exist
	nonExisting := app.poaKeeper.GetValidator(GetContext(app, true), genValidators[len(genValidators)-1].Address)
	require.True(t, nonExisting.Address.Empty())
}

func Test_POAValidatorsReplaceExisting(t *testing.T) {
	app, server := newTestWbApp()
	defer app.CloseConnections()
	defer server.Stop()

	genCoins, err := sdk.ParseCoins("1000000000000000wings")
	require.NoError(t, err)

	genValidators, _, _, genPrivKeys := CreateGenAccounts(8, genCoins)
	curValidators, curPrivKeys := genValidators[:7], genPrivKeys[:7]
	newValidators := genValidators[7:]
	_, err = setGenesis(t, app, curValidators)
	require.NoError(t, err)

	// replace existing with existing validator
	{
		replaceValidator := curValidators[1]
		res := replaceValidators(t, app, curValidators, nil, replaceValidator, curPrivKeys, false)
		CheckResultError(t, poaTypes.ErrValidatorExists(""), res)
	}
	// replace non-existing with existing validator
	{
		oldValidator, replaceValidator := newValidators[0], curValidators[1]
		res := replaceValidators(t, app, curValidators, &oldValidator.Address, replaceValidator, curPrivKeys, false)
		CheckResultError(t, poaTypes.ErrValidatorDoesntExists(""), res)
	}
}

func Test_POAValidatorsMinMaxRange(t *testing.T) {
	t.Parallel()

	defMinValidators, defMaxValidators := poaTypes.DefaultMinValidators, poaTypes.DefaultMaxValidators

	app, server := newTestWbApp()
	defer app.CloseConnections()
	defer server.Stop()

	genCoins, err := sdk.ParseCoins("1000000000000000wings")
	genValidators, _, _, genPrivKeys := CreateGenAccounts(int(defMaxValidators)+1, genCoins)
	curValidators, curPrivKeys := genValidators[:defMaxValidators], genPrivKeys[:defMaxValidators]
	require.NoError(t, err)
	_, err = setGenesis(t, app, curValidators)
	require.NoError(t, err)

	require.Equal(t, defMinValidators, app.poaKeeper.GetMinValidators(GetContext(app, true)))
	require.Equal(t, defMaxValidators, app.poaKeeper.GetMaxValidators(GetContext(app, true)))

	// add (defMaxValidators + 1) validator
	{
		newValidator := genValidators[len(genValidators)-1]
		res := addValidators(t, app, curValidators, []*auth.BaseAccount{newValidator}, curPrivKeys, false)
		CheckResultError(t, poaTypes.ErrMaxValidatorsReached(0), res)
	}

	// remove all validator till defMinValidators is reached
	for len(curValidators) != int(defMinValidators) {
		delValidator := genValidators[len(curValidators)-1]
		removeValidators(t, app, curValidators, []*auth.BaseAccount{delValidator}, curPrivKeys, true)
		curValidators, curPrivKeys = curValidators[:len(curValidators)-1], curPrivKeys[:len(curPrivKeys)-1]
	}

	// remove (defMinValidators - 1) validator
	{
		delValidator := genValidators[len(curValidators)-1]
		res := removeValidators(t, app, curValidators, []*auth.BaseAccount{delValidator}, curPrivKeys, false)
		CheckResultError(t, poaTypes.ErrMinValidatorsReached(0), res)
	}
}

func Test_MultisigVoting(t *testing.T) {
	app, server := newTestWbApp()
	defer app.CloseConnections()
	defer server.Stop()

	genCoins, err := sdk.ParseCoins("1000000000000000wings")
	genValidators, _, _, genPrivKeys := CreateGenAccounts(9, genCoins)
	curValidators, curPrivKeys := genValidators[:7], genPrivKeys[:7]
	require.NoError(t, err)
	_, err = setGenesis(t, app, curValidators)
	require.NoError(t, err)

	nonExistingValidator, nonExistingValidatorPrivKey := genValidators[len(genValidators) - 1], genPrivKeys[len(genPrivKeys) - 1]
	app.accountKeeper.SetAccount(app.NewContext(true, abci.Header{Height: app.LastBlockHeight() + 1}), nonExistingValidator)

	addValidator := genValidators[len(genValidators)-2]
	var callMsgId uint64

	// submit call from non-existing validator
	{
		senderAddr, senderPrivKey := nonExistingValidator.Address, nonExistingValidatorPrivKey
		senderAcc := GetAccountCheckTx(app, senderAddr)
		// create call
		addValidatorMsg := msgspoa.NewMsgAddValidator(addValidator.Address, ethAddresses[0], senderAddr)
		msgID := fmt.Sprintf("addValidator:%s", addValidator.Address)
		submitMsg := msmsg.NewMsgSubmitCall(addValidatorMsg, msgID, senderAddr)
		tx := genTx([]sdk.Msg{submitMsg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
		CheckDeliverSpecificErrorTx(t, app, tx, mstypes.ErrNotValidator(""))
	}

	// submit call for existing validator
	{
		senderAddr, senderPrivKey := curValidators[0].Address, curPrivKeys[0]
		senderAcc := GetAccountCheckTx(app, senderAddr)
		// create call
		addValidatorMsg := msgspoa.NewMsgAddValidator(addValidator.Address, ethAddresses[0], senderAddr)
		msgID := fmt.Sprintf("addValidator:%s", addValidator.Address)
		submitMsg := msmsg.NewMsgSubmitCall(addValidatorMsg, msgID, senderAddr)
		tx := genTx([]sdk.Msg{submitMsg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
		CheckDeliverTx(t, app, tx)
		// check call added
		calls := mstypes.CallsResp{}
		CheckRunQuery(t, app, nil, queryGetCallsPath, &calls)
		require.Equal(t, 1, len(calls))
		callMsgId = calls[0].Call.MsgID
		// check vote added
		require.Equal(t, senderAddr, calls[0].Votes[0])
	}

	// add vote
	{
		senderAddr, senderPrivKey := curValidators[1].Address, curPrivKeys[1]
		senderAcc := GetAccountCheckTx(app, senderAddr)
		// confirm call
		confirmMsg := msmsg.MsgConfirmCall{MsgId: callMsgId, Sender: senderAddr}
		tx := genTx([]sdk.Msg{confirmMsg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
		CheckDeliverTx(t, app, tx)
		// check vote added
		calls := mstypes.CallsResp{}
		CheckRunQuery(t, app, nil, queryGetCallsPath, &calls)
		require.Equal(t, 1, len(calls))
		require.Equal(t, senderAddr, calls[0].Votes[1])
	}

	// vote again (sender has already voted)
	{
		senderAddr, senderPrivKey := curValidators[1].Address, curPrivKeys[1]
		senderAcc := GetAccountCheckTx(app, senderAddr)
		// confirm call
		confirmMsg := msmsg.MsgConfirmCall{MsgId: callMsgId, Sender: senderAddr}
		tx := genTx([]sdk.Msg{confirmMsg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
		CheckDeliverSpecificErrorTx(t, app, tx, mstypes.ErrCallAlreadyApproved(0, ""))
	}

	// revoke confirm (non-existing vote)
	{
		senderAddr, senderPrivKey := curValidators[2].Address, curPrivKeys[2]
		senderAcc := GetAccountCheckTx(app, senderAddr)
		// revoke confirm
		revokeMsg := msmsg.MsgRevokeConfirm{MsgId: callMsgId, Sender: senderAddr}
		tx := genTx([]sdk.Msg{revokeMsg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
		CheckDeliverSpecificErrorTx(t, app, tx, mstypes.ErrCallNotApproved(0, ""))
	}

	// revoke confirm (existing vote)
	{
		senderAddr, senderPrivKey := curValidators[1].Address, curPrivKeys[1]
		senderAcc := GetAccountCheckTx(app, senderAddr)
		// revoke confirm
		revokeMsg := msmsg.MsgRevokeConfirm{MsgId: callMsgId, Sender: senderAddr}
		tx := genTx([]sdk.Msg{revokeMsg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
		CheckDeliverTx(t, app, tx)
		// check vote removed
		calls := mstypes.CallsResp{}
		CheckRunQuery(t, app, nil, queryGetCallsPath, &calls)
		require.Equal(t, 1, len(calls))
		require.NotContains(t, calls[0].Votes, senderAddr)
	}

	// revoke confirm (last vote)
	{
		senderAddr, senderPrivKey := curValidators[0].Address, curPrivKeys[0]
		senderAcc := GetAccountCheckTx(app, senderAddr)
		// revoke confirm
		revokeMsg := msmsg.MsgRevokeConfirm{MsgId: callMsgId, Sender: senderAddr}
		tx := genTx([]sdk.Msg{revokeMsg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
		// TODO: revoking call's last vote doesn't remove the call (panic occurs), is that correct?
		expectedErr := mstypes.ErrWrongCallId(callMsgId)
		require.PanicsWithError(t, expectedErr.Error(), func() {
			DeliverTx(app, tx)
		})

		// check call exists with one vote (last vote was not removed 'cause of the panic)
		calls := mstypes.CallsResp{}
		CheckRunQuery(t, app, nil, queryGetCallsPath, &calls)
		require.Equal(t, 1, len(calls))
		require.Equal(t, 1, len(calls[0].Votes))
	}
}

func Test_MultisigBlockHeight(t *testing.T) {
	app, server := newTestWbApp()
	defer app.CloseConnections()
	defer server.Stop()

	genCoins, err := sdk.ParseCoins("1000000000000000wings")
	genAccs, genAddrs, _, genPrivKeys := CreateGenAccounts(7, genCoins)
	require.NoError(t, err)
	_, err = setGenesis(t, app, genAccs)
	require.NoError(t, err)

	recipientAddr, recipientPrivKey := genAddrs[0], genPrivKeys[0]

	// generate blocks to reach multisig call reject condition
	msIntervalToExecute := app.msKeeper.GetIntervalToExecute(GetContext(app, true))
	blockCountToLimit := int(msIntervalToExecute*2 + 1)
	for curIssueIdx := 0; curIssueIdx < blockCountToLimit; curIssueIdx++ {
		// start block
		app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{ChainID: chainID, Height: app.LastBlockHeight() + 1}})
		// generate submit message
		issueId, msgId := fmt.Sprintf("issue%d", curIssueIdx), strconv.Itoa(curIssueIdx)
		issueMsg := msgs.NewMsgIssueCurrency(currency1Symbol, sdk.NewInt(amount), 0, recipientAddr, issueId)
		submitMsg := msmsg.NewMsgSubmitCall(issueMsg, msgId, recipientAddr)
		// emit transaction
		recepientAcc := GetAccount(app, recipientAddr)
		tx := genTx([]sdk.Msg{submitMsg}, []uint64{recepientAcc.GetAccountNumber()}, []uint64{recepientAcc.GetSequence()}, recipientPrivKey)
		res := app.Deliver(tx)
		require.True(t, res.IsOK(), res.Log)
		// commit block
		app.EndBlock(abci.RequestEndBlock{})
		app.Commit()
	}

	// check rejected calls (request one by one as they are not in queue)
	for i := int64(0); i <= msIntervalToExecute; i++ {
		call := mstypes.CallResp{}
		CheckRunQuery(t, app, mstypes.CallReq{CallId: uint64(i)}, queryGetCallPath, &call)
		require.True(t, call.Call.Rejected)
	}

	// check non-rejected calls (request all as they are in the queue)
	{
		calls := mstypes.CallsResp{}
		CheckRunQuery(t, app, nil, queryGetCallsPath, &calls)
		for _, call := range calls {
			require.False(t, call.Call.Rejected)
		}
	}
	// vote for rejected call
	{
		// prev recipient has already voted, pick a new one
		recipientAddr, recipientPrivKey := genAddrs[1], genPrivKeys[1]
		// pick a rejected callId
		msgId := uint64(0)
		// emit transaction
		confirmMsg := msmsg.NewMsgConfirmCall(msgId, recipientAddr)
		recepientAcc := GetAccountCheckTx(app, recipientAddr)
		tx := genTx([]sdk.Msg{confirmMsg}, []uint64{recepientAcc.GetAccountNumber()}, []uint64{recepientAcc.GetSequence()}, recipientPrivKey)
		CheckDeliverSpecificErrorTx(t, app, tx, mstypes.ErrAlreadyRejected(0))
	}
	// vote for already confirmed call (vote ended)
	{
		// pick a non-rejected once voted callId
		msgId := uint64(blockCountToLimit - 1)
		// vote and approve call
		for i := 1; i < len(genAccs)/2 + 1; i++ {
			recipientAddr, recipientPrivKey := genAddrs[i], genPrivKeys[i]
			confirmMsg := msmsg.NewMsgConfirmCall(msgId, recipientAddr)
			recepientAcc := GetAccountCheckTx(app, recipientAddr)
			tx := genTx([]sdk.Msg{confirmMsg}, []uint64{recepientAcc.GetAccountNumber()}, []uint64{recepientAcc.GetSequence()}, recipientPrivKey)
			CheckDeliverTx(t, app, tx)
		}
		// vote for confirmed call
		{
			i := len(genAddrs) - 1
			recipientAddr, recipientPrivKey := genAddrs[i], genPrivKeys[i]
			confirmMsg := msmsg.NewMsgConfirmCall(msgId, recipientAddr)
			recepientAcc := GetAccountCheckTx(app, recipientAddr)
			tx := genTx([]sdk.Msg{confirmMsg}, []uint64{recepientAcc.GetAccountNumber()}, []uint64{recepientAcc.GetSequence()}, recipientPrivKey)
			CheckDeliverSpecificErrorTx(t, app, tx, mstypes.ErrAlreadyConfirmed(0))
		}
	}
}

func issueCurrencyCheck(t *testing.T, app *WbServiceApp, msgID string, msg msgs.MsgIssueCurrency, recipient sdk.AccAddress,
	genAccs []*auth.BaseAccount, addrs []sdk.AccAddress, privKeys []crypto.PrivKey) {

	// Submit message
	submitMsg := msmsg.NewMsgSubmitCall(msg, msgID, recipient)
	{
		acc := GetAccountCheckTx(app, genAccs[0].Address)
		tx := genTx([]sdk.Msg{submitMsg}, []uint64{acc.GetAccountNumber()}, []uint64{acc.GetSequence()}, privKeys[0])
		CheckDeliverTx(t, app, tx)
	}
	calls := mstypes.CallsResp{}
	CheckRunQuery(t, app, nil, queryGetCallsPath, &calls)
	require.Equal(t, 1, len(calls[0].Votes))

	// Vote, vote, vote...
	confirmMsg := msmsg.MsgConfirmCall{MsgId: calls[0].Call.MsgID}
	for i := 1; i < len(genAccs)/2; i++ {
		{
			confirmMsg.Sender = addrs[i]
			acc := GetAccountCheckTx(app, genAccs[i].Address)
			tx := genTx([]sdk.Msg{confirmMsg}, []uint64{acc.GetAccountNumber()}, []uint64{acc.GetSequence()}, privKeys[i])
			CheckDeliverTx(t, app, tx)
		}
		CheckRunQuery(t, app, nil, queryGetCallsPath, &calls)
		require.Equal(t, i+1, len(calls[0].Votes))
	}

	confirmMsg.Sender = addrs[len(addrs)-1]
	{
		acc := GetAccountCheckTx(app, genAccs[len(addrs)-1].Address)
		tx := genTx([]sdk.Msg{confirmMsg}, []uint64{acc.GetAccountNumber()}, []uint64{acc.GetSequence()}, privKeys[len(addrs)-1])
		CheckDeliverTx(t, app, tx)
	}
	CheckRunQuery(t, app, nil, queryGetCallsPath, &calls)
	require.Equal(t, 0, len(calls))
}

func destroyCurrency(t *testing.T, app *WbServiceApp, msg msgs.MsgDestroyCurrency, genAccs []*auth.BaseAccount, privKeys []crypto.PrivKey) {
	acc := GetAccountCheckTx(app, genAccs[0].Address)
	tx := genTx([]sdk.Msg{msg}, []uint64{acc.GetAccountNumber()}, []uint64{acc.GetSequence()}, privKeys[0])
	CheckDeliverTx(t, app, tx)
}

func checkCurrencyExists(t *testing.T, app *WbServiceApp, symbol string, amount int64, decimals int8) {
	var currency types.Currency
	CheckRunQuery(t, app, types.CurrencyReq{Symbol: currency1Symbol}, queryGetCurrencyPath, &currency)
	require.Equal(t, currency1Symbol, currency.Symbol)
	require.Equal(t, amount, currency.Supply.Int64())
	require.Equal(t, decimals, currency.Decimals)
}

func checkIssueExists(t *testing.T, app *WbServiceApp, issueID string, recipient sdk.AccAddress, amount int64) {
	var issue types.Issue
	CheckRunQuery(t, app, types.IssueReq{IssueID: issueID}, queryGetIssuePath, &issue)
	require.Equal(t, currency1Symbol, issue.Symbol)
	require.Equal(t, amount, issue.Amount.Int64())
	require.Equal(t, recipient, issue.Recipient)
}

func checkReplaceValidators(t *testing.T, app *WbServiceApp, genAccs []*auth.BaseAccount, newValidator *auth.BaseAccount, oldPrivKeys []crypto.PrivKey) {
	privKey := oldPrivKeys[0]
	oldValidators := app.poaKeeper.GetValidators(GetContext(app, true))
	sender := oldValidators[0].Address

	// Submit message
	oldValidator := oldValidators[len(oldValidators)-1].Address
	replaceValidatorMsg := msgspoa.NewMsgReplaceValidator(oldValidator, newValidator.Address, ethAddresses[0], sender)
	acc := GetAccountCheckTx(app, sender)
	msgID := fmt.Sprintf("replaceValidator:%s", newValidator.Address)
	submitMsg := msmsg.NewMsgSubmitCall(replaceValidatorMsg, msgID, sender)
	tx := genTx([]sdk.Msg{submitMsg}, []uint64{acc.GetAccountNumber()}, []uint64{acc.GetSequence()}, privKey)
	CheckDeliverTx(t, app, tx)

	calls := mstypes.CallsResp{}
	CheckRunQuery(t, app, nil, queryGetCallsPath, &calls)
	require.Equal(t, 1, len(calls[0].Votes))
	confirmMsg := msmsg.MsgConfirmCall{MsgId: calls[0].Call.MsgID}
	validatorsAmount := app.poaKeeper.GetValidatorAmount(GetContext(app, true))
	for j, vv := range genAccs[1 : validatorsAmount/2+1] {
		acc := GetAccountCheckTx(app, vv.Address)
		confirmMsg.Sender = vv.Address
		tx := genTx([]sdk.Msg{confirmMsg}, []uint64{acc.GetAccountNumber()}, []uint64{acc.GetSequence()}, oldPrivKeys[j+1])
		CheckDeliverTx(t, app, tx)
	}
	CheckRunQuery(t, app, nil, queryGetCallsPath, &calls)
}

func checkRemoveValidators(t *testing.T, app *WbServiceApp, genAccs []*auth.BaseAccount, rmValidators []*auth.BaseAccount, privKeys []crypto.PrivKey) {
	sender := genAccs[0].Address
	for _, v := range rmValidators {
		// Submit message
		removeValidatorMsg := msgspoa.NewMsgRemoveValidator(v.Address, sender)
		acc := GetAccountCheckTx(app, sender)
		msgID := fmt.Sprintf("removeValidator:%s", v.Address)
		submitMsg := msmsg.NewMsgSubmitCall(removeValidatorMsg, msgID, sender)
		tx := genTx([]sdk.Msg{submitMsg}, []uint64{acc.GetAccountNumber()}, []uint64{acc.GetSequence()}, privKeys[0])
		CheckDeliverTx(t, app, tx)

		calls := mstypes.CallsResp{}
		CheckRunQuery(t, app, nil, queryGetCallsPath, &calls)
		require.Equal(t, 1, len(calls[0].Votes))
		confirmMsg := msmsg.MsgConfirmCall{MsgId: calls[0].Call.MsgID}
		validatorsAmount := app.poaKeeper.GetValidatorAmount(GetContext(app, true))
		for idx, vv := range genAccs[1 : validatorsAmount/2+1] {
			acc := GetAccountCheckTx(app, vv.Address)
			confirmMsg.Sender = vv.Address
			tx := genTx([]sdk.Msg{confirmMsg}, []uint64{acc.GetAccountNumber()}, []uint64{acc.GetSequence()}, privKeys[idx+1])
			CheckDeliverTx(t, app, tx)
		}
	}
}

func checkAddValidators(t *testing.T, app *WbServiceApp, genAccs []*auth.BaseAccount, newValidators []*auth.BaseAccount, privKeys []crypto.PrivKey) {
	sender := genAccs[0].Address
	for _, v := range newValidators {
		// Submit message
		addValidatorMsg := msgspoa.NewMsgAddValidator(v.Address, ethAddresses[0], sender)
		acc := GetAccountCheckTx(app, sender)
		msgID := fmt.Sprintf("addValidator:%s", v.Address)
		submitMsg := msmsg.NewMsgSubmitCall(addValidatorMsg, msgID, sender)
		tx := genTx([]sdk.Msg{submitMsg}, []uint64{acc.GetAccountNumber()}, []uint64{acc.GetSequence()}, privKeys[0])
		CheckDeliverTx(t, app, tx)

		calls := mstypes.CallsResp{}
		CheckRunQuery(t, app, nil, queryGetCallsPath, &calls)
		require.Equal(t, 1, len(calls[0].Votes))
		confirmMsg := msmsg.MsgConfirmCall{MsgId: calls[0].Call.MsgID}
		validatorsAmount := app.poaKeeper.GetValidatorAmount(GetContext(app, true))
		for idx, vv := range genAccs[1 : validatorsAmount/2+1] {
			acc := GetAccountCheckTx(app, vv.Address)
			confirmMsg.Sender = vv.Address
			tx := genTx([]sdk.Msg{confirmMsg}, []uint64{acc.GetAccountNumber()}, []uint64{acc.GetSequence()}, privKeys[idx+1])
			CheckDeliverTx(t, app, tx)
		}
	}
}

func addValidators(t *testing.T, app *WbServiceApp, genAccs []*auth.BaseAccount, newValidators []*auth.BaseAccount, privKeys []crypto.PrivKey, doChecks bool) sdk.Result {
	sender := genAccs[0].Address
	for _, v := range newValidators {
		// Submit message
		addValidatorMsg := msgspoa.NewMsgAddValidator(v.Address, ethAddresses[0], sender)
		acc := GetAccountCheckTx(app, sender)
		msgID := fmt.Sprintf("addValidator:%s", v.Address)
		submitMsg := msmsg.NewMsgSubmitCall(addValidatorMsg, msgID, sender)
		tx := genTx([]sdk.Msg{submitMsg}, []uint64{acc.GetAccountNumber()}, []uint64{acc.GetSequence()}, privKeys[0])
		if doChecks {
			CheckDeliverTx(t, app, tx)
		} else if res := DeliverTx(app, tx); !res.IsOK() {
			return res
		}

		calls := mstypes.CallsResp{}
		CheckRunQuery(t, app, nil, queryGetCallsPath, &calls)
		require.Equal(t, 1, len(calls[0].Votes))
		confirmMsg := msmsg.MsgConfirmCall{MsgId: calls[0].Call.MsgID}
		validatorsAmount := app.poaKeeper.GetValidatorAmount(GetContext(app, true))
		for idx, vv := range genAccs[1 : validatorsAmount/2+1] {
			acc := GetAccountCheckTx(app, vv.Address)
			confirmMsg.Sender = vv.Address
			tx := genTx([]sdk.Msg{confirmMsg}, []uint64{acc.GetAccountNumber()}, []uint64{acc.GetSequence()}, privKeys[idx+1])
			if doChecks {
				CheckDeliverTx(t, app, tx)
			} else if res := DeliverTx(app, tx); !res.IsOK() {
				return res
			}
		}
	}

	return sdk.Result{}
}

func replaceValidators(t *testing.T, app *WbServiceApp, genAccs []*auth.BaseAccount, oldValidatorOverwrite *sdk.AccAddress, newValidator *auth.BaseAccount, oldPrivKeys []crypto.PrivKey, doChecks bool) sdk.Result {
	privKey := oldPrivKeys[0]
	oldValidators := app.poaKeeper.GetValidators(GetContext(app, true))
	sender := oldValidators[0].Address

	// Submit message
	oldValidator := oldValidators[len(oldValidators)-1].Address
	if oldValidatorOverwrite != nil {
		oldValidator = *oldValidatorOverwrite
	}
	replaceValidatorMsg := msgspoa.NewMsgReplaceValidator(oldValidator, newValidator.Address, ethAddresses[0], sender)
	acc := GetAccountCheckTx(app, sender)
	msgID := fmt.Sprintf("replaceValidator:%s", newValidator.Address)
	submitMsg := msmsg.NewMsgSubmitCall(replaceValidatorMsg, msgID, sender)
	tx := genTx([]sdk.Msg{submitMsg}, []uint64{acc.GetAccountNumber()}, []uint64{acc.GetSequence()}, privKey)
	if doChecks {
		CheckDeliverTx(t, app, tx)
	} else if res := DeliverTx(app, tx); !res.IsOK() {
		return res
	}

	calls := mstypes.CallsResp{}
	CheckRunQuery(t, app, nil, queryGetCallsPath, &calls)
	require.Equal(t, 1, len(calls[0].Votes))
	confirmMsg := msmsg.MsgConfirmCall{MsgId: calls[0].Call.MsgID}
	validatorsAmount := app.poaKeeper.GetValidatorAmount(GetContext(app, true))
	for j, vv := range genAccs[1 : validatorsAmount/2+1] {
		acc := GetAccountCheckTx(app, vv.Address)
		confirmMsg.Sender = vv.Address
		tx := genTx([]sdk.Msg{confirmMsg}, []uint64{acc.GetAccountNumber()}, []uint64{acc.GetSequence()}, oldPrivKeys[j+1])
		CheckDeliverTx(t, app, tx)
	}
	if doChecks {
		CheckDeliverTx(t, app, tx)
	} else if res := DeliverTx(app, tx); !res.IsOK() {
		return res
	}

	return sdk.Result{}
}

func removeValidators(t *testing.T, app *WbServiceApp, genAccs []*auth.BaseAccount, rmValidators []*auth.BaseAccount, privKeys []crypto.PrivKey, doChecks bool) sdk.Result {
	sender := genAccs[0].Address
	for _, v := range rmValidators {
		// Submit message
		removeValidatorMsg := msgspoa.NewMsgRemoveValidator(v.Address, sender)
		acc := GetAccountCheckTx(app, sender)
		msgID := fmt.Sprintf("removeValidator:%s", v.Address)
		submitMsg := msmsg.NewMsgSubmitCall(removeValidatorMsg, msgID, sender)
		tx := genTx([]sdk.Msg{submitMsg}, []uint64{acc.GetAccountNumber()}, []uint64{acc.GetSequence()}, privKeys[0])
		if doChecks {
			CheckDeliverTx(t, app, tx)
		} else if res := DeliverTx(app, tx); !res.IsOK() {
			return res
		}

		calls := mstypes.CallsResp{}
		CheckRunQuery(t, app, nil, queryGetCallsPath, &calls)
		require.Equal(t, 1, len(calls[0].Votes))
		confirmMsg := msmsg.MsgConfirmCall{MsgId: calls[0].Call.MsgID}
		validatorsAmount := app.poaKeeper.GetValidatorAmount(GetContext(app, true))
		for idx, vv := range genAccs[1 : validatorsAmount/2+1] {
			acc := GetAccountCheckTx(app, vv.Address)
			confirmMsg.Sender = vv.Address
			tx := genTx([]sdk.Msg{confirmMsg}, []uint64{acc.GetAccountNumber()}, []uint64{acc.GetSequence()}, privKeys[idx+1])
			if doChecks {
				CheckDeliverTx(t, app, tx)
			} else if res := DeliverTx(app, tx); !res.IsOK() {
				return res
			}
		}
	}

	return sdk.Result{}
}
