package simulator

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewUndelegateBondingOp picks a validator and undelegates bonding tokens.
// Undelegation amount = current account delegation amount * {undelegateRatio}.
// Op priorities:
//   validator - highest bonding tokens amount (all statuses);
//   account:
//     - random;
//     - has a validators bonding delegation;
//     - not a validator owner;
func NewUndelegateBondingOp(period time.Duration, undelegateRatio sdk.Dec) *SimOperation {
	checkRatioArg("UndelegateBondingOp", "undelegateRatio", undelegateRatio)

	handler := func(s *Simulator) bool {
		targetAcc, targetVal, udCoin := undelegateOpFindTarget(s, true, undelegateRatio)
		if targetAcc == nil || targetVal == nil {
			return false
		}
		undelegateOpHandle(s, targetAcc, targetVal, udCoin)

		undelegateOpPost(s, targetAcc, targetVal, true)
		s.logger.Info(fmt.Sprintf("UndelegateBondingOp: %s: %s -> %s", targetAcc.Address, targetVal.GetAddress(), s.FormatCoin(udCoin)))

		return true
	}

	return NewSimOperation(period, NewPeriodicNextExecFn(), handler)
}

// NewUndelegateLPOp picks a validator and undelegates LP tokens.
// Undelegation amount = current account delegation amount * {undelegateRatio}.
// Op priorities:
//   validator - highest LP tokens amount (all statuses);
//   account:
//     - random;
//     - has a validators LP delegation;
//     - not a validator owner;
func NewUndelegateLPOp(period time.Duration, undelegateRatio sdk.Dec) *SimOperation {
	checkRatioArg("UndelegateLPOp", "undelegateRatio", undelegateRatio)

	handler := func(s *Simulator) bool {
		targetAcc, targetVal, udCoin := undelegateOpFindTarget(s, false, undelegateRatio)
		if targetAcc == nil || targetVal == nil {
			return false
		}
		undelegateOpHandle(s, targetAcc, targetVal, udCoin)

		undelegateOpPost(s, targetAcc, targetVal, false)
		s.logger.Info(fmt.Sprintf("UndelegateLPOp: %s: %s -> %s", targetAcc.Address, targetVal.GetAddress(), s.FormatCoin(udCoin)))

		return true
	}

	return NewSimOperation(period, NewPeriodicNextExecFn(), handler)
}

func undelegateOpFindTarget(s *Simulator, bondingUD bool, udRatio sdk.Dec) (targetAcc *SimAccount, targetVal *SimValidator, udCoin sdk.Coin) {
	denom := s.stakingDenom
	if !bondingUD {
		denom = s.lpDenom
	}

	// pick a validator with the highest tokens amount (all statuses)
	validators := s.GetAllValidators().GetSortedByTokens(bondingUD, true)
	if len(validators) == 0 {
		return
	}
	targetVal = validators[0]

	for _, acc := range s.GetAllAccounts().GetShuffled() {
		accValAddr := sdk.ValAddress{}
		if acc.IsValOperator() {
			accValAddr = acc.OperatedValidator.GetAddress()
		}

		for _, delegation := range acc.Delegations {
			// check if account did delegate to the selected validator
			if !targetVal.GetAddress().Equals(delegation.ValidatorAddress) {
				continue
			}

			// check not undelegating from the account owned validator
			if accValAddr.Equals(targetVal.GetAddress()) {
				continue
			}

			// estimate undelegation amount
			udAmtDec := sdk.NewDecFromInt(delegation.BondingBalance.Amount)
			if !bondingUD {
				udAmtDec = sdk.NewDecFromInt(delegation.LPBalance.Amount)
			}

			udAmt := udAmtDec.Mul(udRatio).TruncateInt()
			if udAmt.IsZero() {
				continue
			}

			targetAcc = acc
			udCoin = sdk.NewCoin(denom, udAmt)
			return
		}
	}

	return
}

func undelegateOpHandle(s *Simulator, targetAcc *SimAccount, targetVal *SimValidator, udCoin sdk.Coin) {
	s.TxStakeUndelegate(targetAcc, targetVal.GetAddress(), udCoin)
}

func undelegateOpPost(s *Simulator, targetAcc *SimAccount, targetVal *SimValidator, bondingUD bool) {
	// update validator
	s.UpdateValidator(targetVal)
	// update account
	s.UpdateAccount(targetAcc)
	// update stats
	if bondingUD {
		s.counter.BUndelegations++
	} else {
		s.counter.LPUndelegations++
	}

	s.defferQueue.Add(s.prevBlockTime.Add(s.unbondingDur+5*time.Minute), func() {
		s.UpdateAccount(targetAcc)
	})
}
