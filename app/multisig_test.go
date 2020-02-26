package app

import (
	"fmt"
	"github.com/WingsDao/wings-blockchain/x/currencies/msgs"
	msMsgs "github.com/WingsDao/wings-blockchain/x/multisig/msgs"
	msTypes "github.com/WingsDao/wings-blockchain/x/multisig/types"
	poaMsgs "github.com/WingsDao/wings-blockchain/x/poa/msgs"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"strconv"
	"testing"
)

const (
	queryMsGetCallPath   = "/custom/multisig/call"
	queryMsGetCallsPath  = "/custom/multisig/calls"
	queryMsGetCallLastId = "/custom/multisig/lastId"
	queryMsGetUniqueCall = "/custom/multisig/unique"
)

func Test_MSQueries(t *testing.T) {
	app, server := newTestWbApp()
	defer app.CloseConnections()
	defer server.Stop()

	genCoins, err := sdk.ParseCoins("1000000000000000wings")
	require.NoError(t, err)
	genValidators, _, _, _ := CreateGenAccounts(7, genCoins)

	_, err = setGenesis(t, app, genValidators)
	require.NoError(t, err)

	// check call by non-existing uniqueId query
	{
		request := msTypes.UniqueReq{UniqueId: "non-existing-unique-id"}
		CheckRunQuerySpecificError(t, app, request, queryMsGetUniqueCall, msTypes.ErrNotFoundUniqueID(""))
	}
}

func Test_MSVoting(t *testing.T) {
	app, server := newTestWbApp()
	defer app.CloseConnections()
	defer server.Stop()

	genCoins, err := sdk.ParseCoins("1000000000000000wings")
	require.NoError(t, err)
	accs, _, _, privKeys := CreateGenAccounts(9, genCoins)
	genValidators, genPrivKeys := accs[:7], privKeys[:7]
	targetValidator := accs[7]
	nonExistingValidator, nonExistingValidatorPrivKey := accs[8], privKeys[8]

	_, err = setGenesis(t, app, genValidators)
	require.NoError(t, err)

	// create account for non-existing validator
	app.accountKeeper.SetAccount(app.NewContext(true, abci.Header{Height: app.LastBlockHeight() + 1}), nonExistingValidator)

	var callMsgId uint64
	var callUniqueId string

	// submit call from non-existing validator
	{
		// create call
		senderAcc, senderPrivKey := GetAccountCheckTx(app, nonExistingValidator.Address), nonExistingValidatorPrivKey
		addMsg := poaMsgs.NewMsgAddValidator(targetValidator.Address, ethAddresses[0], senderAcc.GetAddress())
		msgID := fmt.Sprintf("addValidator:%s", targetValidator.Address)
		submitMsg := msMsgs.NewMsgSubmitCall(addMsg, msgID, senderAcc.GetAddress())
		tx := genTx([]sdk.Msg{submitMsg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
		CheckDeliverSpecificErrorTx(t, app, tx, msTypes.ErrNotValidator(""))
	}

	// submit call from existing validator
	{
		// create call
		senderAcc, senderPrivKey := GetAccountCheckTx(app, genValidators[0].Address), genPrivKeys[0]
		addMsg := poaMsgs.NewMsgAddValidator(targetValidator.Address, ethAddresses[0], senderAcc.GetAddress())
		callUniqueId = fmt.Sprintf("addValidator:%s", targetValidator.Address)
		submitMsg := msMsgs.NewMsgSubmitCall(addMsg, callUniqueId, senderAcc.GetAddress())
		tx := genTx([]sdk.Msg{submitMsg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
		CheckDeliverTx(t, app, tx)

		// check call added
		calls := msTypes.CallsResp{}
		CheckRunQuery(t, app, nil, queryMsGetCallsPath, &calls)
		require.Equal(t, 1, len(calls))
		callMsgId = calls[0].Call.MsgID
		// check vote added
		require.Equal(t, senderAcc.GetAddress(), calls[0].Votes[0])
	}

	// add vote
	{
		// confirm call
		senderAcc, senderPrivKey := GetAccountCheckTx(app, genValidators[1].Address), genPrivKeys[1]
		confirmMsg := msMsgs.MsgConfirmCall{MsgId: callMsgId, Sender: senderAcc.GetAddress()}
		tx := genTx([]sdk.Msg{confirmMsg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
		CheckDeliverTx(t, app, tx)

		// check vote added
		calls := msTypes.CallsResp{}
		CheckRunQuery(t, app, nil, queryMsGetCallsPath, &calls)
		require.Equal(t, 1, len(calls))
		require.Equal(t, senderAcc.GetAddress(), calls[0].Votes[1])
	}

	// check call lastId query
	{
		response := msTypes.LastIdRes{}
		CheckRunQuery(t, app, nil, queryMsGetCallLastId, &response)
		require.Equal(t, callMsgId, response.LastId)
	}

	// check call by uniqueId query (with votes)
	{
		request := msTypes.UniqueReq{UniqueId: callUniqueId}
		response := msTypes.CallResp{}
		CheckRunQuery(t, app, request, queryMsGetUniqueCall, &response)
		require.Equal(t, callUniqueId, response.Call.UniqueID)
		require.Equal(t, callMsgId, response.Call.MsgID)
		require.Len(t, response.Votes, 2)
		require.ElementsMatch(t, []sdk.AccAddress{genValidators[0].Address, genValidators[1].Address}, response.Votes)
	}

	// vote again (sender has already voted)
	{
		// confirm call
		senderAcc, senderPrivKey := GetAccountCheckTx(app, genValidators[1].Address), genPrivKeys[1]
		confirmMsg := msMsgs.MsgConfirmCall{MsgId: callMsgId, Sender: senderAcc.GetAddress()}
		tx := genTx([]sdk.Msg{confirmMsg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
		CheckDeliverSpecificErrorTx(t, app, tx, msTypes.ErrCallAlreadyApproved(0, ""))
	}

	// revoke confirm (non-existing vote)
	{
		// revoke confirm
		senderAcc, senderPrivKey := GetAccountCheckTx(app, genValidators[2].Address), genPrivKeys[2]
		revokeMsg := msMsgs.MsgRevokeConfirm{MsgId: callMsgId, Sender: senderAcc.GetAddress()}
		tx := genTx([]sdk.Msg{revokeMsg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
		CheckDeliverSpecificErrorTx(t, app, tx, msTypes.ErrCallNotApproved(0, ""))
	}

	// revoke confirm (existing vote)
	{
		// revoke confirm
		senderAcc, senderPrivKey := GetAccountCheckTx(app, genValidators[1].Address), genPrivKeys[1]
		revokeMsg := msMsgs.MsgRevokeConfirm{MsgId: callMsgId, Sender: senderAcc.GetAddress()}
		tx := genTx([]sdk.Msg{revokeMsg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
		CheckDeliverTx(t, app, tx)

		// check vote removed
		calls := msTypes.CallsResp{}
		CheckRunQuery(t, app, nil, queryMsGetCallsPath, &calls)
		require.Equal(t, 1, len(calls))
		require.NotContains(t, calls[0].Votes, senderAcc.GetAddress())
	}

	// revoke confirm (last vote)
	{
		// revoke confirm
		senderAcc, senderPrivKey := GetAccountCheckTx(app, genValidators[0].Address), genPrivKeys[0]
		revokeMsg := msMsgs.MsgRevokeConfirm{MsgId: callMsgId, Sender: senderAcc.GetAddress()}
		tx := genTx([]sdk.Msg{revokeMsg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
		// TODO: revoking call's last vote doesn't remove the call (panic occurs), is that correct?
		expectedErr := msTypes.ErrWrongCallId(callMsgId)
		require.PanicsWithError(t, expectedErr.Error(), func() {
			DeliverTx(app, tx)
		})

		// check call exists with one vote (last vote was not removed 'cause of the panic)
		calls := msTypes.CallsResp{}
		CheckRunQuery(t, app, nil, queryMsGetCallsPath, &calls)
		require.Equal(t, 1, len(calls))
		require.Equal(t, 1, len(calls[0].Votes))
	}
}

func Test_MSBlockHeight(t *testing.T) {
	app, server := newTestWbApp()
	defer app.CloseConnections()
	defer server.Stop()

	genCoins, err := sdk.ParseCoins("1000000000000000wings")
	require.NoError(t, err)
	genAccs, genAddrs, _, genPrivKeys := CreateGenAccounts(7, genCoins)

	_, err = setGenesis(t, app, genAccs)
	require.NoError(t, err)


	// generate blocks to reach multisig call reject condition
	senderAddr, senderPrivKey := genAddrs[0], genPrivKeys[0]
	msIntervalToExecute := app.msKeeper.GetIntervalToExecute(GetContext(app, true))
	blockCountToLimit := int(msIntervalToExecute*2 + 1)
	for curIssueIdx := 0; curIssueIdx < blockCountToLimit; curIssueIdx++ {
		// start block
		app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{ChainID: chainID, Height: app.LastBlockHeight() + 1}})
		// generate submit message
		issueId, msgId := fmt.Sprintf("issue%d", curIssueIdx), strconv.Itoa(curIssueIdx)
		issueMsg := msgs.NewMsgIssueCurrency(currency1Symbol, sdk.NewInt(amount), 0, senderAddr, issueId)
		submitMsg := msMsgs.NewMsgSubmitCall(issueMsg, msgId, senderAddr)
		// emit transaction
		senderAcc := GetAccount(app, senderAddr)
		tx := genTx([]sdk.Msg{submitMsg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
		res := app.Deliver(tx)
		require.True(t, res.IsOK(), res.Log)
		// commit block
		app.EndBlock(abci.RequestEndBlock{})
		app.Commit()
	}

	// check rejected calls (request one by one as they are not in queue)
	for i := int64(0); i <= msIntervalToExecute; i++ {
		call := msTypes.CallResp{}
		CheckRunQuery(t, app, msTypes.CallReq{CallId: uint64(i)}, queryMsGetCallPath, &call)
		require.True(t, call.Call.Rejected)
	}

	// check non-rejected calls (request all as they are in the queue)
	{
		calls := msTypes.CallsResp{}
		CheckRunQuery(t, app, nil, queryMsGetCallsPath, &calls)
		for _, call := range calls {
			require.False(t, call.Call.Rejected)
		}
	}

	// vote for rejected call
	{
		// prev recipient has already voted, pick a new one
		senderAcc, senderPrivKey := GetAccountCheckTx(app, genAddrs[1]), genPrivKeys[1]
		// pick a rejected callId
		msgId := uint64(0)
		// emit transaction
		confirmMsg := msMsgs.NewMsgConfirmCall(msgId, senderAcc.GetAddress())
		tx := genTx([]sdk.Msg{confirmMsg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
		CheckDeliverSpecificErrorTx(t, app, tx, msTypes.ErrAlreadyRejected(0))
	}

	// vote for already confirmed call (vote ended)
	{
		// pick a non-rejected once voted callId
		msgId := uint64(blockCountToLimit - 1)
		// vote and approve call
		for i := 1; i < len(genAccs)/2+1; i++ {
			senderAcc, senderPrivKey := GetAccountCheckTx(app, genAddrs[i]), genPrivKeys[i]
			confirmMsg := msMsgs.NewMsgConfirmCall(msgId, senderAcc.GetAddress())
			tx := genTx([]sdk.Msg{confirmMsg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
			CheckDeliverTx(t, app, tx)
		}

		// vote for confirmed call
		{
			idx := len(genAddrs) - 1
			senderAcc, senderPrivKey := GetAccountCheckTx(app, genAddrs[idx]), genPrivKeys[idx]
			confirmMsg := msMsgs.NewMsgConfirmCall(msgId, senderAcc.GetAddress())
			tx := genTx([]sdk.Msg{confirmMsg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
			CheckDeliverSpecificErrorTx(t, app, tx, msTypes.ErrAlreadyConfirmed(0))
		}
	}

	// revoke rejected call
	{
		// pick a rejected callId
		msgId := uint64(0)
		// emit transaction
		senderAcc, senderPrivKey := GetAccountCheckTx(app, genAddrs[0]), genPrivKeys[0]
		revokeMsg := msMsgs.MsgRevokeConfirm{MsgId: msgId, Sender: senderAcc.GetAddress()}
		tx := genTx([]sdk.Msg{revokeMsg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
		CheckDeliverSpecificErrorTx(t, app, tx, msTypes.ErrAlreadyRejected(0))
	}

	// revoke approved call
	{
		// pick an approved callId
		msgId := uint64(blockCountToLimit - 1)
		// emit transaction
		senderAcc, senderPrivKey := GetAccountCheckTx(app, genAddrs[0]), genPrivKeys[0]
		revokeMsg := msMsgs.MsgRevokeConfirm{MsgId: msgId, Sender: senderAcc.GetAddress()}
		tx := genTx([]sdk.Msg{revokeMsg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
		CheckDeliverSpecificErrorTx(t, app, tx, msTypes.ErrAlreadyConfirmed(0))
	}
}
