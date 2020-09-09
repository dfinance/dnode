package simulator

import (
	"fmt"
	"sort"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/stretchr/testify/require"
)

// NewCreateValidatorOp creates validator for account which is not an operator already.
func NewCreateValidatorOp(period time.Duration) *SimOperation {
	handler := func(s *Simulator) bool {
		// find account without validator
		var simAcc *SimAccount
		for i := 0; i < len(s.accounts); i++ {
			if s.accounts[i].OperatedValidator == nil {
				simAcc = s.accounts[i]
				break
			}
		}

		if simAcc == nil {
			return false
		}

		// define commissions
		comRate, err := sdk.NewDecFromStr("0.100000000000000000")
		require.NoError(s.t, err)

		comMaxRate, err := sdk.NewDecFromStr("0.200000000000000000")
		require.NoError(s.t, err)

		comMaxChangeRate, err := sdk.NewDecFromStr("0.010000000000000000")
		require.NoError(s.t, err)

		// create
		s.TxStakingCreateValidator(simAcc, staking.NewCommissionRates(comRate, comMaxRate, comMaxChangeRate))
		validator := s.QueryStakingValidator(sdk.ValAddress(simAcc.Address))

		// update account
		s.UpdateAccount(simAcc)
		simAcc.OperatedValidator = &validator

		s.logger.Info(fmt.Sprintf("ValidatorOp: created for %s", simAcc.Address))

		return true
	}

	return NewSimOperation(period, NewPeriodicNextExecFn(), handler)
}

// NewDelegateOp picks a validator and searches for account to delegate.
// Validators priority - lowest tokens.
// Accounts priority - didn't delegate to target validator and has enough coins.
func NewDelegateOp(period time.Duration, amount sdk.Coin) *SimOperation {
	handler := func(s *Simulator) bool {
		// sort validators by their tokens (first - min tokens)
		validators := s.GetValidators()
		if len(validators) == 0 {
			return false
		}

		sort.Slice(validators, func(i, j int) bool {
			if validators[i].Tokens.LT(validators[j].Tokens) {
				return true
			}
			return false
		})
		targetVal := validators[0]

		// find an account which didn't delegate to this validator yet and has enough funds
		hasEnoughCoins := func(acc *SimAccount) bool {
			accCoin := acc.Coins.AmountOf(amount.Denom)
			if accCoin.LT(amount.Amount) {
				return false
			}
			return true
		}

		var targetAcc *SimAccount
		for _, acc := range s.accounts {
			if acc.HasDelegation(targetVal.OperatorAddress) {
				continue
			}

			if hasEnoughCoins(acc) {
				targetAcc = acc
				break
			}
		}

		// if account wasn't found, find the first with enough coins
		if targetAcc == nil {
			for _, acc := range s.accounts {
				if hasEnoughCoins(acc) {
					targetAcc = acc
					break
				}
			}
		}

		// not luck check
		if targetAcc == nil {
			return false
		}

		// delegate
		s.TxStakingDelegate(targetAcc, targetVal, amount)

		// update account
		s.UpdateAccount(targetAcc)
		delegation := s.QueryStakingDelegation(targetAcc, targetVal)
		targetAcc.AddDelegation(&delegation)

		// update validator
		s.UpdateValidator(targetVal)

		s.logger.Info(fmt.Sprintf("DelegateOp: delegated %s from %s to %s", amount, targetAcc.Address, targetVal.OperatorAddress))

		return true
	}

	return NewSimOperation(period, NewPeriodicNextExecFn(), handler)
}
