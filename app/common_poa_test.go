package app

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/libs/rand"

	"github.com/dfinance/dnode/x/poa"
)

// AddValidators creates poa add validator multisig message and confirms it.
func AddValidators(t *testing.T, app *DnServiceApp, genAccs []*auth.BaseAccount, newValidators []*auth.BaseAccount, privKeys []crypto.PrivKey, doChecks bool) (*sdk.Result, error) {
	for _, v := range newValidators {
		addMsg := poa.NewMsgAddValidator(v.Address, ethAddresses[0], genAccs[0].Address)
		msgID := fmt.Sprintf("addValidator:%s:%d", v.Address, rand.Uint16())

		res, err := MSMsgSubmitAndVote(t, app, msgID, addMsg, 0, genAccs, privKeys, doChecks)
		if doChecks {
			require.NoError(t, err)
		} else if err != nil {
			return res, err
		}
	}

	return nil, nil
}

// ReplaceValidator creates poa replace validator multisig message and confirms it.
func ReplaceValidator(t *testing.T, app *DnServiceApp, genAccs []*auth.BaseAccount, oldValidatorAddr, newValidatorAddr sdk.AccAddress, oldPrivKeys []crypto.PrivKey, doChecks bool) (*sdk.Result, error) {
	replaceMsg := poa.NewMsgReplaceValidator(oldValidatorAddr, newValidatorAddr, ethAddresses[0], genAccs[0].GetAddress())
	msgID := fmt.Sprintf("ReplaceValidator:%s", newValidatorAddr)

	return MSMsgSubmitAndVote(t, app, msgID, replaceMsg, 0, genAccs, oldPrivKeys, doChecks)
}

// RemoveValidators creates poa remove validator multisig message and confirms it.
func RemoveValidators(t *testing.T, app *DnServiceApp, genAccs []*auth.BaseAccount, rmValidators []*auth.BaseAccount, privKeys []crypto.PrivKey, doChecks bool) (*sdk.Result, error) {
	for _, v := range rmValidators {
		removeMsg := poa.NewMsgRemoveValidator(v.Address, genAccs[0].Address)
		msgID := fmt.Sprintf("removeValidator:%s:%d", v.Address, rand.Uint16())

		res, err := MSMsgSubmitAndVote(t, app, msgID, removeMsg, 0, genAccs, privKeys, doChecks)
		if doChecks {
			require.NoError(t, err)
		} else if err != nil {
			return res, err
		}
	}

	return nil, nil
}
