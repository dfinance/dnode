package app

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"

	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/core/msmodule"
	"github.com/dfinance/dnode/x/multisig"
	msExport "github.com/dfinance/dnode/x/multisig/export"
)

// MSMsgSubmitAndVote submits multi signature message call and confirms it.
func MSMsgSubmitAndVote(t *testing.T, app *DnServiceApp, msMsgID string, msMsg msmodule.MsMsg, submitAccIdx uint, accs []*auth.BaseAccount, privKeys []crypto.PrivKey, doChecks bool) (*sdk.Result, error) {
	confirmCnt := int(app.poaKeeper.GetEnoughConfirmations(GetContext(app, true)))

	// lazy input check
	require.Equal(t, len(accs), len(privKeys), "invalid input: accs / privKeys len mismatch")
	require.Less(t, submitAccIdx, uint(len(accs)), "invalid input: submitAccIdx >= len(accs)")
	require.Less(t, submitAccIdx, uint(len(accs)), "invalid input: submitAccIdx >= len(accs)")
	require.LessOrEqual(t, confirmCnt, len(accs), "invalid input: confirmations count > len(accs)")

	callMsgID := dnTypes.ID{}
	{
		// submit message
		senderAcc, senderPrivKey := GetAccountCheckTx(app, accs[submitAccIdx].Address), privKeys[submitAccIdx]
		submitMsg := msExport.NewMsgSubmitCall(msMsg, msMsgID, senderAcc.GetAddress())
		tx := GenTx([]sdk.Msg{submitMsg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
		if doChecks {
			CheckDeliverTx(t, app, tx)
		} else if res, err := DeliverTx(app, tx); err != nil {
			return res, err
		}

		// check vote added
		calls := multisig.CallsResp{}
		CheckRunQuery(t, app, nil, queryMsGetCallsPath, &calls)
		require.Equal(t, 1, len(calls[0].Votes))

		callMsgID = calls[0].Call.ID
	}

	// cut submit message sender from accounts
	accsFixed, privKeysFixed := append([]*auth.BaseAccount(nil), accs...), append([]crypto.PrivKey(nil), privKeys...)
	accsFixed = append(accsFixed[:submitAccIdx], accsFixed[submitAccIdx+1:]...)
	privKeysFixed = append(privKeysFixed[:submitAccIdx], privKeysFixed[submitAccIdx+1:]...)

	// voting (confirming)
	for idx := 0; idx < confirmCnt-2; idx++ {
		// confirm message
		senderAcc, senderPrivKey := GetAccountCheckTx(app, accsFixed[idx].Address), privKeysFixed[idx]
		confirmMsg := msExport.NewMsgConfirmCall(callMsgID, senderAcc.GetAddress())
		tx := GenTx([]sdk.Msg{confirmMsg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
		if doChecks {
			CheckDeliverTx(t, app, tx)
		} else if res, err := DeliverTx(app, tx); err != nil {
			return res, err
		}

		// check vote added / call removed
		calls := multisig.CallsResp{}
		CheckRunQuery(t, app, nil, queryMsGetCallsPath, &calls)
		require.Equal(t, idx+2, len(calls[0].Votes))
	}

	// voting (last confirm)
	{
		// confirm message
		idx := len(accsFixed) - 1
		senderAcc, senderPrivKey := GetAccountCheckTx(app, accsFixed[idx].Address), privKeysFixed[idx]
		confirmMsg := msExport.NewMsgConfirmCall(callMsgID, senderAcc.GetAddress())
		tx := GenTx([]sdk.Msg{confirmMsg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
		if doChecks {
			CheckDeliverTx(t, app, tx)
		} else if res, err := DeliverTx(app, tx); err != nil {
			return res, err
		}

		// check call removed
		calls := multisig.CallsResp{}
		CheckRunQuery(t, app, nil, queryMsGetCallsPath, &calls)
		require.Equal(t, 0, len(calls))
	}

	return nil, nil
}
