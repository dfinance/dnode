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
			return true
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
		validator := s.QueryStakeValidator(sdk.ValAddress(simAcc.Address))

		// update account
		s.UpdateAccount(simAcc)
		simAcc.OperatedValidator = &validator

		s.logger.Info(fmt.Sprintf("ValidatorOp: created for %s", simAcc.Address))

		return true
	}

	return NewSimOperation(period, NewPeriodicNextExecFn(), handler)
}

// NewDelegateOp picks a validator and searches for account to delegate.
// SelfStake increment is allowed.
// Delegation amount = current account balance * ratioCoef.
// Op priorities:
//   validator - lowest tokens amount;
//   account - random, enough coins;
func NewDelegateOp(period time.Duration, delegateRatio sdk.Dec) *SimOperation {
	checkRatioArg("DelegateOp", "delegateRatio", delegateRatio)

	handler := func(s *Simulator) bool {
		// pick a validator with the lowest tokens amount
		validators := s.GetValidatorSortedByStake(false)
		if len(validators) == 0 {
			return false
		}
		validator := validators[0]

		// pick a target account with enough coins
		var delAmt sdk.Int
		var targetAcc *SimAccount
		for _, acc := range s.GetShuffledAccounts() {
			// estimate delegation amount
			accCoinAmtDec := sdk.NewDecFromInt(acc.Coins.AmountOf(s.stakingDenom))
			delAmt = accCoinAmtDec.Mul(delegateRatio).TruncateInt()
			if delAmt.IsZero() {
				continue
			}

			targetAcc = acc
		}
		if targetAcc == nil {
			return false
		}

		// delegate
		coin := sdk.NewCoin(s.stakingDenom, delAmt)
		s.TxStakingDelegate(targetAcc, validator, coin)

		// update account
		s.UpdateAccount(targetAcc)
		// update validator
		s.UpdateValidator(validator)
		// update stats
		s.counter.Delegations++

		s.logger.Info(fmt.Sprintf("DelegateOp: %s: %s -> %s", targetAcc.Address, delAmt, validator.OperatorAddress))

		return true
	}

	return NewSimOperation(period, NewPeriodicNextExecFn(), handler)
}

// NewRedelegateOp picks a validator and redelegate tokens to an other validator.
// Redelegation amount = current account delegation amount * ratioCoef.
// Op priorities:
//   dstValidator - lowest tokens amount;
//   srcValidator - highest account delegation shares;
//   account:
//     - random;
//     - has no active redelegations with srcValidator and dstValidator;
//     - has enough coins;
//     - not a dstValidator owner;
func NewRedelegateOp(period time.Duration, redelegateRatio sdk.Dec) *SimOperation {
	checkRatioArg("RedelegateOp", "redelegateRatio", redelegateRatio)

	handler := func(s *Simulator) bool {
		// pick a dstValidator with the lowest tokens amount
		validators := s.GetValidatorSortedByStake(false)
		if len(validators) == 0 {
			return false
		}
		dstValidator := validators[0]

		rdInProcess := func(accAddr sdk.AccAddress, srcValAddr, dstValAddr sdk.ValAddress) bool {
			for _, rd := range s.QueryStakeRedelegations(accAddr, sdk.ValAddress{}, sdk.ValAddress{}) {
				if rd.ValidatorSrcAddress.Equals(srcValAddr) || rd.ValidatorDstAddress.Equals(srcValAddr) {
					return true
				}

				if rd.ValidatorSrcAddress.Equals(dstValAddr) || rd.ValidatorDstAddress.Equals(dstValAddr) {
					return true
				}

				return false
			}

			return false
		}

		// pick a target account
		for _, acc := range s.GetShuffledAccounts() {
			accValAddr := sdk.ValAddress{}
			if acc.OperatedValidator != nil {
				accValAddr = acc.OperatedValidator.OperatorAddress
			}

			// check not redelegating to the account owned validator
			if dstValidator.OperatorAddress.Equals(accValAddr) {
				continue
			}

			// pick a delegation with the highest share
			for _, delegation := range GetSortedDelegation(acc.Delegations, true) {
				srcValidator := s.GetValidatorByAddress(delegation.ValidatorAddress)

				if srcValidator.OperatorAddress.Equals(dstValidator.OperatorAddress) {
					continue
				}

				// check not redelegating from the account owned validator
				if srcValidator.OperatorAddress.Equals(accValAddr) {
					continue
				}

				// check if an other redelegation is in progress for the selected account
				if rdInProcess(acc.Address, srcValidator.OperatorAddress, dstValidator.OperatorAddress) {
					continue
				}

				// estimate redelegation amount
				rdAmtDec := sdk.NewDecFromInt(delegation.Balance.Amount)
				rdAmt := rdAmtDec.Mul(redelegateRatio).TruncateInt()
				if rdAmt.IsZero() {
					continue
				}

				// redelegate
				coin := sdk.NewCoin(delegation.Balance.Denom, rdAmt)
				s.TxStakingRedelegate(acc, srcValidator.OperatorAddress, dstValidator.OperatorAddress, coin)

				// update validators
				s.UpdateValidator(srcValidator)
				s.UpdateValidator(dstValidator)
				// update account
				s.UpdateAccount(acc)
				// update stats
				s.counter.Redelegations++

				s.logger.Info(fmt.Sprintf("RedelegateOp: %s: %s -> %s -> %s", acc.Address, srcValidator.OperatorAddress, rdAmt, dstValidator.OperatorAddress))

				return true
			}
		}

		return true
	}

	return NewSimOperation(period, NewPeriodicNextExecFn(), handler)
}

// NewUndelegateOp picks a validator and undelegates tokens.
// Undelegation amount = current account delegation amount * ratioCoef.
// Op priorities:
//   validator - highest tokens amount;
//   account:
//     - random;
//     - has a validators delegation;
//     - not a validator owner;
func NewUndelegateOp(period time.Duration, undelegateRatio sdk.Dec) *SimOperation {
	checkRatioArg("UndelegateOp", "undelegateRatio", undelegateRatio)

	handler := func(s *Simulator) bool {
		// pick a validator with the highest tokens amount;
		validators := s.GetValidatorSortedByStake(true)
		if len(validators) == 0 {
			return false
		}
		validator := validators[0]

		for _, acc := range s.GetShuffledAccounts() {
			accValAddr := sdk.ValAddress{}
			if acc.OperatedValidator != nil {
				accValAddr = acc.OperatedValidator.OperatorAddress
			}

			for _, delegation := range acc.Delegations {
				// check if account did delegate to the selected validator
				if !validator.OperatorAddress.Equals(delegation.ValidatorAddress) {
					continue
				}

				// check not undelegating from the account owned validator
				if accValAddr.Equals(validator.OperatorAddress) {
					continue
				}

				// estimate undelegation amount
				udAmtDec := sdk.NewDecFromInt(delegation.Balance.Amount)
				udAmt := udAmtDec.Mul(undelegateRatio).TruncateInt()
				if udAmt.IsZero() {
					continue
				}

				// undelegate
				coin := sdk.NewCoin(delegation.Balance.Denom, udAmt)
				s.TxStakingUndelegate(acc, validator.OperatorAddress, coin)

				// update validator
				s.UpdateValidator(validator)
				// update account
				s.UpdateAccount(acc)
				// update stats
				s.counter.Undelegations++

				s.defferQueue.Add(s.prevBlockTime.Add(s.unbondingDur+5*time.Minute), func() {
					s.UpdateAccount(acc)
				})

				s.logger.Info(fmt.Sprintf("UndelegateOp: %s: %s -> %s", acc.Address, validator.OperatorAddress, udAmt))

				return true
			}
		}

		return false
	}

	return NewSimOperation(period, NewPeriodicNextExecFn(), handler)
}

// NewGetDelRewardOp take delegator rewards.
// Op priority:
//   account;
//     - random;
//     - has delegations;
//   validator - random account delegation;
func NewGetDelRewardOp(period time.Duration) *SimOperation {
	handler := func(s *Simulator) bool {
		for _, acc := range s.GetShuffledAccounts() {
			if len(acc.Delegations) == 0 {
				continue
			}
			targetDelegation := GetShuffledDelegations(acc.Delegations)[0]

			rewardsDec := s.QueryDistDelReward(acc.Address, targetDelegation.ValidatorAddress)
			rewards := rewardsDec.AmountOf(s.stakingDenom).TruncateInt()

			// distribute
			s.TxDistributionReward(acc, targetDelegation.ValidatorAddress)

			// update account
			s.UpdateAccount(acc)
			// update stats
			s.counter.Rewards++
			s.counter.RewardsCollected = s.counter.RewardsCollected.Add(rewards)

			s.logger.Info(fmt.Sprintf("DelRewardOp: %s from %s: %s", acc.Address, targetDelegation.ValidatorAddress, rewards))

			return true
		}

		return false
	}

	return NewSimOperation(period, NewPeriodicNextExecFn(), handler)
}

// NewGetValRewardOp takes validator commissions rewards.
// Op priority:
//   validator - random;
func NewGetValRewardOp(period time.Duration) *SimOperation {
	handler := func(s *Simulator) bool {
		for _, acc := range s.GetShuffledAccounts() {
			if acc.OperatedValidator == nil {
				continue
			}

			rewardsDec := s.QueryDistrValCommission(acc.OperatedValidator.OperatorAddress)
			rewards := rewardsDec.AmountOf(s.stakingDenom).TruncateInt()

			// distribute
			s.TxDistributionCommission(acc, acc.OperatedValidator.OperatorAddress)

			// update account
			s.UpdateAccount(acc)
			// update stats
			s.counter.Commissions++
			s.counter.CommissionsCollected = s.counter.CommissionsCollected.Add(rewards)

			s.logger.Info(fmt.Sprintf("ValRewardOp: %s for %s: %s", acc.OperatedValidator.OperatorAddress, acc.Address, rewards))

			return true
		}

		return false
	}

	return NewSimOperation(period, NewPeriodicNextExecFn(), handler)
}

func checkRatioArg(opName, argName string, argValue sdk.Dec) {
	errMsgPrefix := fmt.Sprintf("%s: %s: ", opName, argName)
	if argValue.LTE(sdk.ZeroDec()) {
		panic(fmt.Errorf("%s: LTE 0", errMsgPrefix))
	}
	if argValue.GT(sdk.OneDec()) {
		panic(fmt.Errorf("%s: GE 1", errMsgPrefix))
	}
}
