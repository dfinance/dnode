// +build unit

package app

import (
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/libs/rand"

	"github.com/dfinance/dnode/x/core"
	"github.com/dfinance/dnode/x/poa"
)

const (
	queryPoaGetValidatorsPath   = "/custom/" + poa.ModuleName + "/" + poa.QueryValidators
	queryPoaGetValidatorPath    = "/custom/" + poa.ModuleName + "/" + poa.QueryValidator
	queryPoaGetMinMaxParamsPath = "/custom/" + poa.ModuleName + "/" + poa.QueryMinMax
)

// Checks that poa module supports only multisig calls for validator msgs (using MSRouter).
func TestPOAApp_HandlerIsMultisigOnly(t *testing.T) {
	t.Parallel()

	app, server := newTestDnApp()
	defer app.CloseConnections()
	defer server.Stop()

	accs, _, _, privKeys := CreateGenAccounts(8, GenDefCoins(t))
	genValidators, genPrivKeys, newValidators := accs[:7], privKeys[:7], accs[7:]

	_, err := setGenesis(t, app, genValidators)
	require.NoError(t, err)

	// check module supports only multisig calls (using MSRouter)
	{
		senderAcc, senderPrivKey := GetAccountCheckTx(app, genValidators[0].Address), genPrivKeys[0]
		addMsg := poa.NewMsgAddValidator(newValidators[0].Address, ethAddresses[0], genValidators[0].Address)
		tx := genTx([]sdk.Msg{addMsg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
		CheckDeliverSpecificErrorTx(t, app, tx, core.ErrNotMultisigModule)
	}
}

// Test poa module queries.
func TestPOAApp_Queries(t *testing.T) {
	t.Parallel()

	app, server := newTestDnApp()
	defer app.CloseConnections()
	defer server.Stop()

	genValidators, _, _, _ := CreateGenAccounts(7, GenDefCoins(t))

	_, err := setGenesis(t, app, genValidators)
	require.NoError(t, err)

	validators := app.poaKeeper.GetValidators(GetContext(app, true))

	// getAllValidators query check
	{
		response := poa.ValidatorsConfirmationsResp{}
		CheckRunQuery(t, app, nil, queryPoaGetValidatorsPath, &response)
		require.Equal(t, validators, response.Validators)
		require.Equal(t, uint16(len(response.Validators)/2+1), response.Confirmations)
	}

	// getValidator query check
	{
		reqValidator := validators[0]
		request := poa.ValidatorReq{Address: reqValidator.Address}
		response := poa.Validator{}
		CheckRunQuery(t, app, request, queryPoaGetValidatorPath, &response)

		require.Equal(t, reqValidator.Address, response.Address)
		require.Equal(t, reqValidator.EthAddress, response.EthAddress)
	}

	// check minMax query
	{
		response := poa.Params{}
		CheckRunQuery(t, app, nil, queryPoaGetMinMaxParamsPath, &response)
		require.Equal(t, poa.DefaultMinValidators, response.MinValidators)
		require.Equal(t, poa.DefaultMaxValidators, response.MaxValidators)
	}
}

// Test adding validator.
func TestPOAApp_ValidatorsAdd(t *testing.T) {
	t.Parallel()

	app, server := newTestDnApp()
	defer app.CloseConnections()
	defer server.Stop()

	accs, _, _, privKeys := CreateGenAccounts(11, GenDefCoins(t))
	genValidators, genPrivKeys, newValidators := accs[:7], privKeys[:7], accs[7:]

	_, err := setGenesis(t, app, genValidators)
	require.NoError(t, err)

	// add new validators
	curConfirmCnt := app.poaKeeper.GetEnoughConfirmations(GetContext(app, true))
	{
		addValidators(t, app, genValidators, newValidators, genPrivKeys, true)

		added := 0
		validators := app.poaKeeper.GetValidators(GetContext(app, true))
	Loop:
		for _, v := range newValidators {
			for _, vv := range validators {
				if v.Address.String() == vv.Address.String() {
					added++
					continue Loop
				}
			}
		}
		require.Equal(t, added, len(newValidators))
		require.Equal(t, len(newValidators)+len(genValidators), len(validators))
	}

	// check hasValidator helper function
	{
		for _, v := range newValidators {
			require.True(t, app.poaKeeper.HasValidator(GetContext(app, true), v.Address))
		}
	}

	// check confirmation count increased
	{
		newConfirmCnt := app.poaKeeper.GetEnoughConfirmations(GetContext(app, true))
		require.Greater(t, newConfirmCnt, curConfirmCnt)
		curConfirmCnt = newConfirmCnt

	}

	// add already existing validator
	{
		res, err := addValidators(t, app, genValidators, []*auth.BaseAccount{newValidators[0]}, genPrivKeys, false)
		CheckResultError(t, poa.ErrValidatorExists, res, err)
	}
}

// Test removing validator.
func TestPOAApp_ValidatorsRemove(t *testing.T) {
	t.Parallel()

	app, server := newTestDnApp()
	defer app.CloseConnections()
	defer server.Stop()

	accs, _, _, privKeys := CreateGenAccounts(11, GenDefCoins(t))
	genValidators, genPrivKeys, targetValidators := accs[:7], privKeys[:7], accs[7:]

	_, err := setGenesis(t, app, genValidators)
	require.NoError(t, err)

	// add validators to remove later
	addValidators(t, app, genValidators, targetValidators, genPrivKeys, true)
	require.Equal(t, len(genValidators)+len(targetValidators), int(app.poaKeeper.GetValidatorAmount(GetContext(app, true))))
	curConfirmCnt := app.poaKeeper.GetEnoughConfirmations(GetContext(app, true))

	// remove validators
	{
		removeValidators(t, app, genValidators, targetValidators, genPrivKeys, true)
		require.Equal(t, len(genValidators), int(app.poaKeeper.GetValidatorAmount(GetContext(app, true))))

		// check requested rmValidators were removed
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
	}

	// check hasValidator helper function
	{
		for _, v := range targetValidators {
			require.False(t, app.poaKeeper.HasValidator(GetContext(app, true), v.Address))
		}
	}

	// check confirmation count decreased
	{
		newConfirmCnt := app.poaKeeper.GetEnoughConfirmations(GetContext(app, true))
		require.Less(t, newConfirmCnt, curConfirmCnt)
		curConfirmCnt = newConfirmCnt

	}

	// remove non-existing validator
	{
		res, err := removeValidators(t, app, genValidators, []*auth.BaseAccount{targetValidators[0]}, genPrivKeys, false)
		CheckResultError(t, poa.ErrValidatorNotExists, res, err)
	}
}

// Test replacing validator.
func TestPOAApp_ValidatorsReplace(t *testing.T) {
	t.Parallel()

	app, server := newTestDnApp()
	defer app.CloseConnections()
	defer server.Stop()

	accs, _, _, privKeys := CreateGenAccounts(8, GenDefCoins(t))
	genValidators, genPrivKeys, targetValidators := accs[:7], privKeys[:7], accs[7:]

	_, err := setGenesis(t, app, genValidators)
	require.NoError(t, err)

	oldValidator, newValidator := genValidators[len(genValidators)-1], targetValidators[0]

	// replace
	{
		replaceValidator(t, app, genValidators, oldValidator.Address, newValidator.Address, genPrivKeys, true)
	}

	// check "new" validator was added ("old" replaced)
	{
		rcvValidator, err := app.poaKeeper.GetValidator(GetContext(app, true), newValidator.Address)
		require.NoError(t, err)
		require.Equal(t, newValidator.Address.String(), rcvValidator.Address.String())
		require.Equal(t, len(genValidators), int(app.poaKeeper.GetValidatorAmount(GetContext(app, true))))
	}

	// check "old" validator doesn't exist
	{
		_, err := app.poaKeeper.GetValidator(GetContext(app, true), oldValidator.Address)
		require.True(t, poa.ErrValidatorNotExists.Is(err))
	}
}

// Test replacing existed validator.
func TestPOAApp_ValidatorsReplaceExisting(t *testing.T) {
	t.Parallel()

	app, server := newTestDnApp()
	defer app.CloseConnections()
	defer server.Stop()

	accs, _, _, privKeys := CreateGenAccounts(8, GenDefCoins(t))
	genValidators, genPrivKeys, targetValidators := accs[:7], privKeys[:7], accs[7:]

	_, err := setGenesis(t, app, genValidators)
	require.NoError(t, err)

	// replace existing with existing validator
	{
		oldValidator, newValidator := genValidators[0], genValidators[1]
		res, err := replaceValidator(t, app, genValidators, oldValidator.Address, newValidator.Address, genPrivKeys, false)
		CheckResultError(t, poa.ErrValidatorExists, res, err)
	}

	// replace non-existing with existing validator
	{
		nonExistingValidator, newValidator := targetValidators[0], genValidators[1]
		res, err := replaceValidator(t, app, genValidators, nonExistingValidator.Address, newValidator.Address, genPrivKeys, false)
		CheckResultError(t, poa.ErrValidatorNotExists, res, err)
	}
}

// Test
func TestPOAApp_MinMaxRange(t *testing.T) {
	t.Parallel()

	app, server := newTestDnApp()
	defer app.CloseConnections()
	defer server.Stop()

	defMinValidators, defMaxValidators := poa.DefaultMinValidators, poa.DefaultMaxValidators
	accs, _, _, privKeys := CreateGenAccounts(int(defMaxValidators)+1, GenDefCoins(t))
	genValidators, genPrivKeys, targetValidators := accs[:defMaxValidators], privKeys[:defMaxValidators], accs[defMaxValidators:]

	_, err := setGenesis(t, app, genValidators)
	require.NoError(t, err)

	// check module params are set to default values
	require.Equal(t, defMinValidators, app.poaKeeper.GetMinValidators(GetContext(app, true)))
	require.Equal(t, defMaxValidators, app.poaKeeper.GetMaxValidators(GetContext(app, true)))

	// check adding (defMaxValidators + 1) validator
	{
		newValidator := targetValidators[0]
		res, err := addValidators(t, app, genValidators, []*auth.BaseAccount{newValidator}, genPrivKeys, false)
		CheckResultError(t, poa.ErrMaxValidatorsReached, res, err)
	}

	// check removing (defMinValidators - 1) validator
	{
		// remove all validator till defMinValidators is reached
		curValidators, curPrivKeys := genValidators, genPrivKeys
		for len(curValidators) != int(defMinValidators) {
			delValidator := curValidators[len(curValidators)-1]
			removeValidators(t, app, curValidators, []*auth.BaseAccount{delValidator}, curPrivKeys, true)
			curValidators, curPrivKeys = curValidators[:len(curValidators)-1], curPrivKeys[:len(curPrivKeys)-1]
		}

		// remove the last one
		delValidator := genValidators[len(curValidators)-1]
		res, err := removeValidators(t, app, curValidators, []*auth.BaseAccount{delValidator}, curPrivKeys, false)
		CheckResultError(t, poa.ErrMinValidatorsReached, res, err)
	}
}

// addValidators creates poa add validator multisig message and confirms it.
func addValidators(t *testing.T, app *DnServiceApp, genAccs []*auth.BaseAccount, newValidators []*auth.BaseAccount, privKeys []crypto.PrivKey, doChecks bool) (*sdk.Result, error) {
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

// replaceValidator creates poa replace validator multisig message and confirms it.
func replaceValidator(t *testing.T, app *DnServiceApp, genAccs []*auth.BaseAccount, oldValidatorAddr, newValidatorAddr sdk.AccAddress, oldPrivKeys []crypto.PrivKey, doChecks bool) (*sdk.Result, error) {
	replaceMsg := poa.NewMsgReplaceValidator(oldValidatorAddr, newValidatorAddr, ethAddresses[0], genAccs[0].GetAddress())
	msgID := fmt.Sprintf("replaceValidator:%s", newValidatorAddr)

	return MSMsgSubmitAndVote(t, app, msgID, replaceMsg, 0, genAccs, oldPrivKeys, doChecks)
}

// removeValidators creates poa remove validator multisig message and confirms it.
func removeValidators(t *testing.T, app *DnServiceApp, genAccs []*auth.BaseAccount, rmValidators []*auth.BaseAccount, privKeys []crypto.PrivKey, doChecks bool) (*sdk.Result, error) {
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
