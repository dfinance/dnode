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
	id := "UndelegateBondingOp"
	checkRatioArg(id, "undelegateRatio", undelegateRatio)

	handler := func(s *Simulator) (bool, string) {
		targetAcc, targetVal, udCoin := undelegateOpFindTarget(s, true, undelegateRatio)
		if targetAcc == nil || targetVal == nil {
			return false, "target not found"
		}
		undelegateOpHandle(s, targetAcc, targetVal, udCoin)

		undelegateOpPost(s, targetAcc, targetVal, true)
		msg := fmt.Sprintf("%s: %s -> %s", targetAcc.Address, targetVal.GetAddress(), s.FormatCoin(udCoin))

		return true, msg
	}

	return NewSimOperation(id, period, NewPeriodicNextExecFn(), handler)
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
	id := "UndelegateLPOp"
	checkRatioArg(id, "undelegateRatio", undelegateRatio)

	handler := func(s *Simulator) (bool, string) {
		targetAcc, targetVal, udCoin := undelegateOpFindTarget(s, false, undelegateRatio)
		if targetAcc == nil || targetVal == nil {
			return false, "target not found"
		}
		undelegateOpHandle(s, targetAcc, targetVal, udCoin)

		undelegateOpPost(s, targetAcc, targetVal, false)
		msg := fmt.Sprintf("%s: %s -> %s", targetAcc.Address, targetVal.GetAddress(), s.FormatCoin(udCoin))

		return true, msg
	}

	return NewSimOperation(id, period, NewPeriodicNextExecFn(), handler)
}

func undelegateOpFindTarget(s *Simulator, bondingUD bool, udRatio sdk.Dec) (targetAcc *SimAccount, targetVal *SimValidator, udCoin sdk.Coin) {
	denom := s.stakingDenom
	if !bondingUD {
		denom = s.lpDenom
	}

	// pick a validator with the highest tokens amount (all statuses)
	vals := s.GetAllValidators().GetSortedByTokens(bondingUD, true)
	for _, val := range vals {
		// pick a random account
		accs := s.GetAllAccounts().GetShuffled()
		for _, acc := range accs {
			accValAddr := sdk.ValAddress{}
			if acc.IsValOperator() {
				accValAddr = acc.OperatedValidator.GetAddress()
			}

			// pick a corresponding delegation (targetValidator)
			for _, delegation := range acc.Delegations {
				// check if account did delegate to the selected validator
				if !val.GetAddress().Equals(delegation.ValidatorAddress) {
					continue
				}

				// check not undelegating from the account owned validator
				if accValAddr.Equals(val.GetAddress()) {
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
				targetVal = val
				udCoin = sdk.NewCoin(denom, udAmt)
				return
			}
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
