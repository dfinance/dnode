package app

import (
	"fmt"
	msTypes "github.com/WingsDao/wings-blockchain/x/multisig/types"
	msgsPoa "github.com/WingsDao/wings-blockchain/x/poa/msgs"
	poaTypes "github.com/WingsDao/wings-blockchain/x/poa/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"
	"testing"
)

const (
	queryPoaGetValidatorsPath   = "/custom/poa/validators"
	queryPoaGetValidatorPath    = "/custom/poa/validator"
	queryPoaGetMinMaxParamsPath = "/custom/poa/minmax"
)

func Test_POAHandlerIsMultisigOnly(t *testing.T) {
	app, server := newTestWbApp()
	defer app.CloseConnections()
	defer server.Stop()

	genCoins, err := sdk.ParseCoins("1000000000000000wings")
	require.NoError(t, err)
	accs, _, _, privKeys := CreateGenAccounts(8, genCoins)
	genValidators, genPrivKeys, newValidators := accs[:7], privKeys[:7], accs[7:]

	_, err = setGenesis(t, app, genValidators)
	require.NoError(t, err)

	// check module supports only multisig calls (using MSRouter)
	{
		senderAcc, senderPrivKey := GetAccountCheckTx(app, genValidators[0].Address), genPrivKeys[0]
		addMsg := msgsPoa.NewMsgAddValidator(newValidators[0].Address, ethAddresses[0], genValidators[0].Address)
		tx := genTx([]sdk.Msg{addMsg}, []uint64{senderAcc.GetAccountNumber()}, []uint64{senderAcc.GetSequence()}, senderPrivKey)
		CheckDeliverSpecificErrorTx(t, app, tx, msTypes.ErrOnlyMultisig(poaTypes.DefaultCodespace, poaTypes.ModuleName))
	}
}

func Test_POAQueries(t *testing.T) {
	app, server := newTestWbApp()
	defer app.CloseConnections()
	defer server.Stop()

	genCoins, err := sdk.ParseCoins("1000000000000000wings")
	require.NoError(t, err)
	genValidators, _, _, _ := CreateGenAccounts(7, genCoins)

	_, err = setGenesis(t, app, genValidators)
	require.NoError(t, err)

	validators := app.poaKeeper.GetValidators(GetContext(app, true))

	// getAllValidators query check
	{
		response := poaTypes.ValidatorsConfirmations{}
		CheckRunQuery(t, app, nil, queryPoaGetValidatorsPath, &response)
		require.Equal(t, validators, response.Validators)
		require.Equal(t, uint16(len(response.Validators)/2+1), response.Confirmations)
	}

	// getValidator query check
	{
		reqValidator := validators[0]
		request := poaTypes.QueryValidator{Address: reqValidator.Address}
		response := poaTypes.Validator{}
		CheckRunQuery(t, app, request, queryPoaGetValidatorPath, &response)

		require.Equal(t, reqValidator.Address, response.Address)
		require.Equal(t, reqValidator.EthAddress, response.EthAddress)
	}

	// check minMax query
	{
		response := poaTypes.Params{}
		CheckRunQuery(t, app, nil, queryPoaGetMinMaxParamsPath, &response)
		require.Equal(t, poaTypes.DefaultMinValidators, response.MinValidators)
		require.Equal(t, poaTypes.DefaultMaxValidators, response.MaxValidators)
	}
}

func Test_POAValidatorsAdd(t *testing.T) {
	app, server := newTestWbApp()
	defer app.CloseConnections()
	defer server.Stop()

	genCoins, err := sdk.ParseCoins("1000000000000000wings")
	require.NoError(t, err)
	accs, _, _, privKeys := CreateGenAccounts(11, genCoins)
	genValidators, genPrivKeys, newValidators := accs[:7], privKeys[:7], accs[7:]

	_, err = setGenesis(t, app, genValidators)
	require.NoError(t, err)

	// add new validators
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

	// add already existing validator
	{
		res := addValidators(t, app, genValidators, []*auth.BaseAccount{newValidators[0]}, genPrivKeys, false)
		CheckResultError(t, poaTypes.ErrValidatorExists(""), res)
	}
}

func Test_POAValidatorsRemove(t *testing.T) {
	app, server := newTestWbApp()
	defer app.CloseConnections()
	defer server.Stop()

	genCoins, err := sdk.ParseCoins("1000000000000000wings")
	require.NoError(t, err)
	accs, _, _, privKeys := CreateGenAccounts(11, genCoins)
	genValidators, genPrivKeys, targetValidators := accs[:7], privKeys[:7], accs[7:]

	_, err = setGenesis(t, app, genValidators)
	require.NoError(t, err)

	// add validators to remove later
	addValidators(t, app, genValidators, targetValidators, genPrivKeys, true)
	require.Equal(t, len(genValidators)+len(targetValidators), int(app.poaKeeper.GetValidatorAmount(GetContext(app, true))))

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

	// remove non-existing validator
	{
		res := removeValidators(t, app, genValidators, []*auth.BaseAccount{targetValidators[0]}, genPrivKeys, false)
		CheckResultError(t, poaTypes.ErrValidatorDoesntExists(""), res)
	}
}

func Test_POAValidatorsReplace(t *testing.T) {
	app, server := newTestWbApp()
	defer app.CloseConnections()
	defer server.Stop()

	genCoins, err := sdk.ParseCoins("1000000000000000wings")
	require.NoError(t, err)
	accs, _, _, privKeys := CreateGenAccounts(8, genCoins)
	genValidators, genPrivKeys, targetValidators := accs[:7], privKeys[:7], accs[7:]

	_, err = setGenesis(t, app, genValidators)
	require.NoError(t, err)

	oldValidator, newValidator := genValidators[len(genValidators)-1], targetValidators[0]

	// replace
	{
		replaceValidator(t, app, genValidators, oldValidator.Address, newValidator.Address, genPrivKeys, true)
	}

	// check "new" validator was added ("old" replaced)
	{
		rcvValidator := app.poaKeeper.GetValidator(GetContext(app, true), newValidator.Address)
		require.Equal(t, newValidator.Address.String(), rcvValidator.Address.String())
		require.Equal(t, len(genValidators), int(app.poaKeeper.GetValidatorAmount(GetContext(app, true))))
	}

	// check "old" validator doesn't exist
	{
		nonExistingValidator := app.poaKeeper.GetValidator(GetContext(app, true), oldValidator.Address)
		require.True(t, nonExistingValidator.Address.Empty())
	}
}

func Test_POAValidatorsReplaceExisting(t *testing.T) {
	app, server := newTestWbApp()
	defer app.CloseConnections()
	defer server.Stop()

	genCoins, err := sdk.ParseCoins("1000000000000000wings")
	require.NoError(t, err)
	accs, _, _, privKeys := CreateGenAccounts(8, genCoins)
	genValidators, genPrivKeys, targetValidators := accs[:7], privKeys[:7], accs[7:]

	_, err = setGenesis(t, app, genValidators)
	require.NoError(t, err)

	// replace existing with existing validator
	{
		oldValidator, newValidator := genValidators[0], genValidators[1]
		res := replaceValidator(t, app, genValidators, oldValidator.Address, newValidator.Address, genPrivKeys, false)
		CheckResultError(t, poaTypes.ErrValidatorExists(""), res)
	}
	// replace non-existing with existing validator
	{
		nonExistingValidator, newValidator := targetValidators[0], genValidators[1]
		res := replaceValidator(t, app, genValidators, nonExistingValidator.Address, newValidator.Address, genPrivKeys, false)
		CheckResultError(t, poaTypes.ErrValidatorDoesntExists(""), res)
	}
}

func Test_POAValidatorsMinMaxRange(t *testing.T) {
	defMinValidators, defMaxValidators := poaTypes.DefaultMinValidators, poaTypes.DefaultMaxValidators

	app, server := newTestWbApp()
	defer app.CloseConnections()
	defer server.Stop()

	genCoins, err := sdk.ParseCoins("1000000000000000wings")
	require.NoError(t, err)
	accs, _, _, privKeys := CreateGenAccounts(int(defMaxValidators)+1, genCoins)
	genValidators, genPrivKeys, targetValidators := accs[:defMaxValidators], privKeys[:defMaxValidators], accs[defMaxValidators:]

	_, err = setGenesis(t, app, genValidators)
	require.NoError(t, err)

	// check module params are set to default values
	require.Equal(t, defMinValidators, app.poaKeeper.GetMinValidators(GetContext(app, true)))
	require.Equal(t, defMaxValidators, app.poaKeeper.GetMaxValidators(GetContext(app, true)))

	// check adding (defMaxValidators + 1) validator
	{
		newValidator := targetValidators[0]
		res := addValidators(t, app, genValidators, []*auth.BaseAccount{newValidator}, genPrivKeys, false)
		CheckResultError(t, poaTypes.ErrMaxValidatorsReached(0), res)
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
		res := removeValidators(t, app, curValidators, []*auth.BaseAccount{delValidator}, curPrivKeys, false)
		CheckResultError(t, poaTypes.ErrMinValidatorsReached(0), res)
	}
}

func addValidators(t *testing.T, app *WbServiceApp, genAccs []*auth.BaseAccount, newValidators []*auth.BaseAccount, privKeys []crypto.PrivKey, doChecks bool) sdk.Result {
	for _, v := range newValidators {
		addMsg := msgsPoa.NewMsgAddValidator(v.Address, ethAddresses[0], genAccs[0].Address)
		msgID := fmt.Sprintf("addValidator:%s", v.Address)

		res := MSMsgSubmitAndVote(t, app, msgID, addMsg, 0, genAccs, privKeys, doChecks)
		if doChecks {
			require.True(t, res.IsOK())
		} else if !res.IsOK() {
			return res
		}
	}

	return sdk.Result{}
}

func replaceValidator(t *testing.T, app *WbServiceApp, genAccs []*auth.BaseAccount, oldValidatorAddr, newValidatorAddr sdk.AccAddress, oldPrivKeys []crypto.PrivKey, doChecks bool) sdk.Result {
	replaceMsg := msgsPoa.NewMsgReplaceValidator(oldValidatorAddr, newValidatorAddr, ethAddresses[0], genAccs[0].GetAddress())
	msgID := fmt.Sprintf("replaceValidator:%s", newValidatorAddr)

	res := MSMsgSubmitAndVote(t, app, msgID, replaceMsg, 0, genAccs, oldPrivKeys, doChecks)
	if doChecks {
		require.True(t, res.IsOK())
	}

	return res
}

func removeValidators(t *testing.T, app *WbServiceApp, genAccs []*auth.BaseAccount, rmValidators []*auth.BaseAccount, privKeys []crypto.PrivKey, doChecks bool) sdk.Result {
	for _, v := range rmValidators {
		removeMsg := msgsPoa.NewMsgRemoveValidator(v.Address, genAccs[0].Address)
		msgID := fmt.Sprintf("removeValidator:%s", v.Address)

		res := MSMsgSubmitAndVote(t, app, msgID, removeMsg, 0, genAccs, privKeys, doChecks)
		if doChecks {
			require.True(t, res.IsOK())
		} else if !res.IsOK() {
			return res
		}
	}

	return sdk.Result{}
}
