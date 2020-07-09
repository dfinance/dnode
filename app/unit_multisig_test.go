// +build unit

package app

import (
	"fmt"
	"strconv"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"

	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/currencies"
	"github.com/dfinance/dnode/x/multisig"
	msExport "github.com/dfinance/dnode/x/multisig/export"
	"github.com/dfinance/dnode/x/poa"
)

// Test multisig module queries.
func TestMSApp_Queries(t *testing.T) {
	t.Parallel()

	app, appStop := NewTestDnAppMockVM()
	defer appStop()

	genValidators, _, _, _ := CreateGenAccounts(7, GenDefCoins(t))
	CheckSetGenesisMockVM(t, app, genValidators)

	// check call by non-existing uniqueID query
	{
		request := multisig.CallByUniqueIdReq{UniqueID: "non-existing-unique-id"}
		CheckRunQuerySpecificError(t, app, request, queryMsGetUniqueCall, multisig.ErrWrongCallUniqueId)
	}
}

// Multisig votings scenarios.
func TestMSApp_Voting(t *testing.T) {
	t.Parallel()

	app, appStop := NewTestDnAppMockVM()
	defer appStop()

	genValidators, _, _, genPrivKeys := CreateGenAccounts(9, GenDefCoins(t))
	CheckSetGenesisMockVM(t, app, genValidators)

	targetValidator := genValidators[7]
	nonExistingValidator, nonExistingValidatorPrivKey := genValidators[8], genPrivKeys[8]

	// remove validators making them nonExistingValidator for further test cases
	RemoveValidators(t, app, genValidators, []*auth.BaseAccount{targetValidator, nonExistingValidator}, genPrivKeys, true)
	genValidators, genPrivKeys = genValidators[:7], genPrivKeys[:7]

	var callMsgId dnTypes.ID
	var callUniqueId string

	// submit call (from non-existing validator)
	{
		// create call
		senderAcc, senderPrivKey := GetAccountCheckTx(app, nonExistingValidator.Address), nonExistingValidatorPrivKey
		addMsg := poa.NewMsgAddValidator(targetValidator.Address, ethAddresses[0], senderAcc.GetAddress())
		msgID := fmt.Sprintf("addValidator:%s", targetValidator.Address)
		submitMsg := msExport.NewMsgSubmitCall(addMsg, msgID, senderAcc.GetAddress())
		tx := GenTx([]sdk.Msg{submitMsg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
		CheckDeliverSpecificErrorTx(t, app, tx, multisig.ErrPoaNotValidator)
	}

	// submit call (from existing validator)
	{
		// create call
		senderAcc, senderPrivKey := GetAccountCheckTx(app, genValidators[0].Address), genPrivKeys[0]
		addMsg := poa.NewMsgAddValidator(targetValidator.Address, ethAddresses[0], senderAcc.GetAddress())
		callUniqueId = fmt.Sprintf("addValidator:%s", targetValidator.Address)
		submitMsg := msExport.NewMsgSubmitCall(addMsg, callUniqueId, senderAcc.GetAddress())
		tx := GenTx([]sdk.Msg{submitMsg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
		CheckDeliverTx(t, app, tx)

		// check call added
		calls := multisig.CallsResp{}
		CheckRunQuery(t, app, nil, queryMsGetCallsPath, &calls)
		require.Equal(t, 1, len(calls))
		callMsgId = calls[0].Call.ID
		// check vote added
		require.Equal(t, senderAcc.GetAddress(), calls[0].Votes[0])
	}

	// add vote
	{
		// confirm call
		senderAcc, senderPrivKey := GetAccountCheckTx(app, genValidators[1].Address), genPrivKeys[1]
		confirmMsg := msExport.MsgConfirmCall{CallID: callMsgId, Sender: senderAcc.GetAddress()}
		tx := GenTx([]sdk.Msg{confirmMsg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
		CheckDeliverTx(t, app, tx)

		// check vote added
		calls := multisig.CallsResp{}
		CheckRunQuery(t, app, nil, queryMsGetCallsPath, &calls)
		require.Equal(t, 1, len(calls))
		require.Equal(t, senderAcc.GetAddress(), calls[0].Votes[1])
	}

	// check call lastID query
	{
		response := multisig.LastCallIdResp{}
		CheckRunQuery(t, app, nil, queryMsGetCallLastId, &response)
		require.Equal(t, callMsgId.UInt64(), response.LastID.UInt64())
	}

	// check call by uniqueID query (with votes)
	{
		request := multisig.CallByUniqueIdReq{UniqueID: callUniqueId}
		response := multisig.CallResp{}
		CheckRunQuery(t, app, request, queryMsGetUniqueCall, &response)
		require.Equal(t, callUniqueId, response.Call.UniqueID)
		require.Equal(t, callMsgId.UInt64(), response.Call.ID.UInt64())
		require.Len(t, response.Votes, 2)
		require.ElementsMatch(t, []sdk.AccAddress{genValidators[0].Address, genValidators[1].Address}, response.Votes)
	}

	// vote again (sender has already voted)
	{
		// confirm call
		senderAcc, senderPrivKey := GetAccountCheckTx(app, genValidators[1].Address), genPrivKeys[1]
		confirmMsg := msExport.MsgConfirmCall{CallID: callMsgId, Sender: senderAcc.GetAddress()}
		tx := GenTx([]sdk.Msg{confirmMsg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
		CheckDeliverSpecificErrorTx(t, app, tx, multisig.ErrVoteAlreadyConfirmed)
	}

	// vote (from non-existing validator)
	{
		// confirm call
		senderAcc, senderPrivKey := GetAccountCheckTx(app, nonExistingValidator.Address), nonExistingValidatorPrivKey
		confirmMsg := msExport.MsgConfirmCall{CallID: callMsgId, Sender: senderAcc.GetAddress()}
		tx := GenTx([]sdk.Msg{confirmMsg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
		CheckDeliverSpecificErrorTx(t, app, tx, multisig.ErrPoaNotValidator)
	}

	// revoke confirm (from non-existing validator)
	{
		// revoke confirm
		senderAcc, senderPrivKey := GetAccountCheckTx(app, nonExistingValidator.Address), nonExistingValidatorPrivKey
		revokeMsg := msExport.MsgRevokeConfirm{CallID: callMsgId, Sender: senderAcc.GetAddress()}
		tx := GenTx([]sdk.Msg{revokeMsg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
		CheckDeliverErrorTx(t, app, tx)
	}

	// revoke confirm (non-existing vote)
	{
		// revoke confirm
		senderAcc, senderPrivKey := GetAccountCheckTx(app, genValidators[2].Address), genPrivKeys[2]
		revokeMsg := msExport.MsgRevokeConfirm{CallID: callMsgId, Sender: senderAcc.GetAddress()}
		tx := GenTx([]sdk.Msg{revokeMsg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
		CheckDeliverSpecificErrorTx(t, app, tx, multisig.ErrVoteNotApproved)
	}

	// revoke confirm (existing vote)
	{
		// revoke confirm
		senderAcc, senderPrivKey := GetAccountCheckTx(app, genValidators[1].Address), genPrivKeys[1]
		revokeMsg := msExport.MsgRevokeConfirm{CallID: callMsgId, Sender: senderAcc.GetAddress()}
		tx := GenTx([]sdk.Msg{revokeMsg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
		CheckDeliverTx(t, app, tx)

		// check vote removed
		calls := multisig.CallsResp{}
		CheckRunQuery(t, app, nil, queryMsGetCallsPath, &calls)
		require.Equal(t, 1, len(calls))
		require.NotContains(t, calls[0].Votes, senderAcc.GetAddress())
	}

	// revoke confirm (last vote)
	{
		// revoke confirm
		senderAcc, senderPrivKey := GetAccountCheckTx(app, genValidators[0].Address), genPrivKeys[0]
		revokeMsg := msExport.MsgRevokeConfirm{CallID: callMsgId, Sender: senderAcc.GetAddress()}
		tx := GenTx([]sdk.Msg{revokeMsg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
		CheckDeliverTx(t, app, tx)

		// check call revoked
		calls := multisig.CallsResp{}
		CheckRunQuery(t, app, nil, queryMsGetCallsPath, &calls)
		require.Empty(t, calls)
	}
}

// Check calls rejecting with blockHeight timeouts.
func TestMSApp_BlockHeight(t *testing.T) {
	t.Parallel()

	app, appStop := NewTestDnAppMockVM()
	defer appStop()

	genAccs, genAddrs, _, genPrivKeys := CreateGenAccounts(7, GenDefCoins(t))
	CheckSetGenesisMockVM(t, app, genAccs)

	CreateCurrency(t, app, currency1Denom, 0)

	// generate blocks to reach multisig call reject condition
	senderAddr, senderPrivKey := genAddrs[0], genPrivKeys[0]
	msIntervalToExecute := app.msKeeper.GetIntervalToExecute(GetContext(app, true))
	blockCountToLimit := int(msIntervalToExecute*2 + 1)
	for curIssueIdx := 0; curIssueIdx < blockCountToLimit; curIssueIdx++ {
		// start block
		app.BeginBlock(abci.RequestBeginBlock{Header: abci.Header{ChainID: chainID, Height: app.LastBlockHeight() + 1}})
		// generate submit message
		issueId, msgId := fmt.Sprintf("issue%d", curIssueIdx), strconv.Itoa(curIssueIdx)
		issueMsg := currencies.NewMsgIssueCurrency(issueId, coin1, senderAddr)
		submitMsg := msExport.NewMsgSubmitCall(issueMsg, msgId, senderAddr)
		// emit transaction
		senderAcc := GetAccount(app, senderAddr)
		tx := GenTx([]sdk.Msg{submitMsg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
		_, res, err := app.Deliver(tx)
		require.NoError(t, err, ResultErrorMsg(res, err))
		// commit block
		app.EndBlock(abci.RequestEndBlock{})
		app.Commit()
	}

	// check rejected calls (request one by one as they are not in queue)
	for i := int64(0); i <= msIntervalToExecute; i++ {
		call := multisig.CallResp{}
		CheckRunQuery(t, app, multisig.CallReq{CallID: dnTypes.NewIDFromUint64(uint64(i))}, queryMsGetCallPath, &call)
		require.True(t, call.Call.Rejected)
	}

	// check non-rejected calls (request all as they are in the queue)
	{
		calls := multisig.CallsResp{}
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
		msgId := dnTypes.NewIDFromUint64(0)
		// emit transaction
		confirmMsg := msExport.NewMsgConfirmCall(msgId, senderAcc.GetAddress())
		tx := GenTx([]sdk.Msg{confirmMsg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
		CheckDeliverSpecificErrorTx(t, app, tx, multisig.ErrVoteAlreadyRejected)
	}

	// vote for already confirmed call (vote ended)
	{
		// pick a non-rejected once voted callId
		msgId := dnTypes.NewIDFromUint64(uint64(blockCountToLimit - 1))
		// vote and approve call
		for i := 1; i < len(genAccs)/2+1; i++ {
			senderAcc, senderPrivKey := GetAccountCheckTx(app, genAddrs[i]), genPrivKeys[i]
			confirmMsg := msExport.NewMsgConfirmCall(msgId, senderAcc.GetAddress())
			tx := GenTx([]sdk.Msg{confirmMsg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
			CheckDeliverTx(t, app, tx)
		}

		// vote for approved call
		{
			idx := len(genAddrs) - 1
			senderAcc, senderPrivKey := GetAccountCheckTx(app, genAddrs[idx]), genPrivKeys[idx]
			confirmMsg := msExport.NewMsgConfirmCall(msgId, senderAcc.GetAddress())
			tx := GenTx([]sdk.Msg{confirmMsg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
			CheckDeliverSpecificErrorTx(t, app, tx, multisig.ErrVoteAlreadyApproved)
		}
	}

	// revoke rejected call
	{
		// pick a rejected callId
		msgId := dnTypes.NewIDFromUint64(0)
		// emit transaction
		senderAcc, senderPrivKey := GetAccountCheckTx(app, genAddrs[0]), genPrivKeys[0]
		revokeMsg := msExport.MsgRevokeConfirm{CallID: msgId, Sender: senderAcc.GetAddress()}
		tx := GenTx([]sdk.Msg{revokeMsg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
		CheckDeliverSpecificErrorTx(t, app, tx, multisig.ErrVoteAlreadyRejected)
	}

	// revoke approved call
	{
		// pick an approved callId
		msgId := dnTypes.NewIDFromUint64(uint64(blockCountToLimit - 1))
		// emit transaction
		senderAcc, senderPrivKey := GetAccountCheckTx(app, genAddrs[0]), genPrivKeys[0]
		revokeMsg := msExport.MsgRevokeConfirm{CallID: msgId, Sender: senderAcc.GetAddress()}
		tx := GenTx([]sdk.Msg{revokeMsg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
		CheckDeliverSpecificErrorTx(t, app, tx, multisig.ErrVoteAlreadyApproved)
	}
}
