package simulator

import (
	"fmt"
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
		validators := s.GetValidatorSortedByStake(false)
		if len(validators) == 0 {
			return false
		}
		validator := validators[0]

		var targetAcc *SimAccount
		accList := s.GetShuffledAccounts()
		for _, acc := range accList {
			if acc.HasDelegation(validator.OperatorAddress) {
				continue
			}

			if acc.HasEnoughCoins(amount) {
				targetAcc = acc
				break
			}
		}

		// if account wasn't found, find the first with enough coins
		if targetAcc == nil {
			for _, acc := range accList {
				if acc.HasEnoughCoins(amount) {
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
		s.TxStakingDelegate(targetAcc, validator, amount)

		// update account
		s.UpdateAccount(targetAcc)
		delegation := s.QueryStakingDelegation(targetAcc, validator)
		targetAcc.AddDelegation(&delegation)

		// update validator
		s.UpdateValidator(validator)

		s.logger.Info(fmt.Sprintf("DelegateOp: delegated %s from %s to %s", amount, targetAcc.Address, validator.OperatorAddress))

		return true
	}

	return NewSimOperation(period, NewPeriodicNextExecFn(), handler)
}

// NewRedelegateOp picks a validator and redelegate tokens to other validator.
func NewRedelegateOp(period time.Duration) *SimOperation {
	handler := func(s *Simulator) bool {
		accList := s.GetShuffledAccounts()

		// find dstValidator with the minimal stake
		validators := s.GetValidatorSortedByStake(false)
		if len(validators) == 0 {
			return false
		}
		dstValidator := validators[0]

		rdlInProcess := s.QueryAllRedelegations()
		isInProcess := func(signer sdk.Address, src, dst sdk.ValAddress) bool {
			for _, rp := range rdlInProcess {
				if rp.DelegatorAddress.Equals(signer) && rp.ValidatorDstAddress.Equals(src) {
					return true
				}

				if rp.ValidatorDstAddress.Equals(dst) && rp.ValidatorSrcAddress.Equals(src) {
					return true
				}

				if rp.ValidatorDstAddress.Equals(src) && rp.ValidatorSrcAddress.Equals(dst) {
					return true
				}
			}

			return false
		}

		for _, acc := range accList {
			delegations := GetSortedDelegation(s.QueryAccountDelegations(acc.Address), true)
			// trying find max delegations to the foreign dstValidator
			for _, delegation := range delegations {
				if !dstValidator.OperatorAddress.Equals(delegation.ValidatorAddress) {
					if isInProcess(acc.Address, delegation.ValidatorAddress, dstValidator.OperatorAddress) {
						continue
					}

					srcValidator := s.GetValidatorByAddress(delegation.ValidatorAddress)

					redelegationAmount := delegation.Balance.Amount.Quo(sdk.NewIntFromUint64(2))
					rdCoin := sdk.NewCoin(delegation.Balance.Denom, redelegationAmount)

					s.TxStakingRedelegate(acc, srcValidator.OperatorAddress, dstValidator.OperatorAddress, rdCoin)
					s.UpdateValidator(srcValidator)
					s.UpdateValidator(dstValidator)

					return true
				}
			}
		}

		return true
	}

	return NewSimOperation(period, NewPeriodicNextExecFn(), handler)
}

// NewUndelegateOp picks a validator and redelegate tokens to other validator.
func NewUndelegateOp(period time.Duration) *SimOperation {
	handler := func(s *Simulator) bool {
		accList := s.GetShuffledAccounts()

		validators := s.GetValidatorSortedByStake(false)
		if len(validators) == 0 {
			return false
		}
		validator := validators[0]

		for _, acc := range accList {
			delegations := GetSortedDelegation(s.QueryAccountDelegations(acc.Address), true)
			for _, delegation := range delegations {
				if !validator.OperatorAddress.Equals(delegation.ValidatorAddress) {
					if s.QueryHasUndelegation(acc.Address, delegation.ValidatorAddress) {
						continue
					}

					srcValidator := s.GetValidatorByAddress(delegation.ValidatorAddress)

					unstakeAmount := delegation.Balance.Amount.Quo(sdk.NewIntFromUint64(2))
					rdCoin := sdk.NewCoin(delegation.Balance.Denom, unstakeAmount)

					s.TxStakingUndelegate(acc, srcValidator.OperatorAddress, rdCoin)
					s.UpdateValidator(srcValidator)

					return true
				}
			}
		}

		return false
	}

	return NewSimOperation(period, NewPeriodicNextExecFn(), handler)
}

// NewUndelegateOp picks a validator and redelegate tokens to other validator.
func NewTakeReward(period time.Duration) *SimOperation {
	handler := func(s *Simulator) bool {
		accList := s.GetShuffledAccounts()

		validators := s.GetValidatorSortedByStake(false)
		if len(validators) == 0 {
			return false
		}
		validator := validators[0]

		for _, acc := range accList {
			delegations := GetSortedDelegation(s.QueryAccountDelegations(acc.Address), true)
			for _, delegation := range delegations {
				if !validator.OperatorAddress.Equals(delegation.ValidatorAddress) {
					if s.QueryHasUndelegation(acc.Address, delegation.ValidatorAddress) {
						continue
					}

					srcValidator := s.GetValidatorByAddress(delegation.ValidatorAddress)

					unstakeAmount := delegation.Balance.Amount.Quo(sdk.NewIntFromUint64(2))
					rdCoin := sdk.NewCoin(delegation.Balance.Denom, unstakeAmount)

					s.TxStakingUndelegate(acc, srcValidator.OperatorAddress, rdCoin)
					s.UpdateValidator(srcValidator)

					return true
				}
			}
		}

		return false
	}

	return NewSimOperation(period, NewPeriodicNextExecFn(), handler)
}
