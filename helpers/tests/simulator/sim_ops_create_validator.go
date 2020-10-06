package simulator

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/stretchr/testify/require"
)

// NewCreateValidatorOp creates validator for an account which is not an operator yet and has enough coins.
func NewCreateValidatorOp(period time.Duration, maxValidators uint) *SimOperation {
	id := "ValidatorOp"
	handler := func(s *Simulator) (bool, string) {
		if createValidatorOpCheckInput(s, maxValidators) {
			return true, ""
		}

		targetAcc := createValidatorOpFindTarget(s)
		if targetAcc == nil {
			return true, "target not found"
		}
		createValidatorOpHandle(s, targetAcc)

		createdVal := createValidatorOpPost(s, targetAcc)
		msg := fmt.Sprintf("%s (%s) created for %s", createdVal.GetAddress(), createdVal.Validator.GetConsAddr(), targetAcc.Address)

		return true, msg
	}

	return NewSimOperation(id, period, NewPeriodicNextExecFn(), handler)
}

func createValidatorOpCheckInput(s *Simulator, maxValidators uint) (stop bool) {
	// check limit is reached
	if len(s.GetAllValidators()) >= int(maxValidators) {
		return true
	}

	return false
}

func createValidatorOpFindTarget(s *Simulator) (targetAcc *SimAccount) {
	selfDelegationAmt := s.minSelfDelegationLvl

	// pick an account without a validator
	for _, acc := range s.GetAllAccounts().GetShuffled() {
		if acc.IsValOperator() {
			continue
		}

		// check balance
		if acc.Coins.AmountOf(s.stakingDenom).LT(selfDelegationAmt) {
			continue
		}

		targetAcc = acc
		break
	}

	return
}

func createValidatorOpHandle(s *Simulator, targetAcc *SimAccount) {
	// define commissions
	comRate, err := sdk.NewDecFromStr("0.100000000000000000")
	require.NoError(s.t, err)

	comMaxRate, err := sdk.NewDecFromStr("0.200000000000000000")
	require.NoError(s.t, err)

	comMaxChangeRate, err := sdk.NewDecFromStr("0.010000000000000000")
	require.NoError(s.t, err)

	// create a new validator with min self-delegation
	s.TxStakeCreateValidator(targetAcc, staking.NewCommissionRates(comRate, comMaxRate, comMaxChangeRate))
	s.beginBlock()
	s.endBlock()
}

func createValidatorOpPost(s *Simulator, targetAcc *SimAccount) (createdVal *SimValidator) {
	// update account
	validator := s.QueryStakeValidator(sdk.ValAddress(targetAcc.Address))
	s.UpdateAccount(targetAcc)
	targetAcc.OperatedValidator = NewSimValidator(validator)

	return targetAcc.OperatedValidator
}
