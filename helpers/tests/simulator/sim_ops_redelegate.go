package simulator

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewRedelegateBondingOp picks a validator and redelegate bonding tokens to an other validator.
// Redelegation amount = current account delegation amount * {redelegateRatio}.
// Op priorities:
//   dstValidator:
//     - bonded;
//     - lowest bonding tokens amount;
//   srcValidator - highest account delegation bonding shares;
//   account:
//     - random;
//     - has no active redelegations with srcValidator and dstValidator;
//     - has enough bonding coins;
//     - not a dstValidator owner;
func NewRedelegateBondingOp(period time.Duration, redelegateRatio sdk.Dec) *SimOperation {
	id := "RedelegateBondingOp"
	checkRatioArg(id, "redelegateRatio", redelegateRatio)

	handler := func(s *Simulator) (bool, string) {
		targetAcc, srcValidator, dstValidator, rdCoin := redelegateOpFindTarget(s, true, redelegateRatio)
		if srcValidator == nil || dstValidator == nil {
			return false, "target not found"
		}
		redelegateOpHandle(s, targetAcc, srcValidator, dstValidator, rdCoin)

		redelegateOpPost(s, targetAcc, srcValidator, dstValidator, true)
		msg := fmt.Sprintf("%s: %s -> %s -> %s", targetAcc.Address, srcValidator.GetAddress(), s.FormatCoin(rdCoin), dstValidator.GetAddress())

		return true, msg
	}

	return NewSimOperation(id, period, NewPeriodicNextExecFn(), handler)
}

// NewRedelegateLPOp picks a validator and redelegate LP tokens to an other validator.
// Redelegation amount = current account delegation amount * {redelegateRatio}.
// Op priorities:
//   dstValidator:
//     - bonded;
//     - lowest LP tokens amount;
//   srcValidator - highest account delegation LP shares;
//   account:
//     - random;
//     - has no active redelegations with srcValidator and dstValidator;
//     - has enough LP coins;
//     - not a dstValidator owner;
func NewRedelegateLPOp(period time.Duration, redelegateRatio sdk.Dec) *SimOperation {
	id := "RedelegateLPOp"
	checkRatioArg(id, "redelegateRatio", redelegateRatio)

	handler := func(s *Simulator) (bool, string) {
		targetAcc, srcValidator, dstValidator, rdCoin := redelegateOpFindTarget(s, false, redelegateRatio)
		if srcValidator == nil || dstValidator == nil {
			return false, "target not found"
		}
		redelegateOpHandle(s, targetAcc, srcValidator, dstValidator, rdCoin)

		redelegateOpPost(s, targetAcc, srcValidator, dstValidator, false)
		msg := fmt.Sprintf("%s: %s -> %s -> %s", targetAcc.Address, srcValidator.GetAddress(), s.FormatCoin(rdCoin), dstValidator.GetAddress())

		return true, msg
	}

	return NewSimOperation(id, period, NewPeriodicNextExecFn(), handler)
}

func redelegateOpFindTarget(s *Simulator, bondingRD bool, rdRatio sdk.Dec) (targetAcc *SimAccount, srcValidator, dstValidator *SimValidator, rdCoin sdk.Coin) {
	denom := s.stakingDenom
	if !bondingRD {
		denom = s.lpDenom
	}

	// pick a bonded dstValidator with the lowest tokens amount
	validators := s.GetValidators(true, false, false).GetSortedByTokens(bondingRD, false)
	if len(validators) == 0 {
		return
	}
	dstValidator = validators[0]

	rdInProgress := func(accAddr sdk.AccAddress, srcValAddr, dstValAddr sdk.ValAddress) bool {
		for _, rd := range s.QueryStakeRedelegations(accAddr, sdk.ValAddress{}, sdk.ValAddress{}) {
			if rd.ValidatorSrcAddress.Equals(srcValAddr) || rd.ValidatorDstAddress.Equals(srcValAddr) {
				return true
			}

			if rd.ValidatorSrcAddress.Equals(dstValAddr) || rd.ValidatorDstAddress.Equals(dstValAddr) {
				return true
			}
		}
		return false
	}

	// pick a target account
	accs := s.GetAllAccounts().GetShuffled()
	for _, acc := range accs {
		accValAddr := sdk.ValAddress{}
		if acc.IsValOperator() {
			accValAddr = acc.OperatedValidator.GetAddress()
		}

		// check not redelegating to the account owned validator
		if dstValidator.GetAddress().Equals(accValAddr) {
			continue
		}

		// pick a delegation with the highest share
		delegations := acc.GetSortedDelegations(bondingRD, true)
		for _, delegation := range delegations {
			srcValidatorApplicant := validators.GetByAddress(delegation.ValidatorAddress)

			// check if applicant was found (that validator can be unbonded by now)
			if srcValidatorApplicant == nil {
				continue
			}

			// check not the one picked above
			if srcValidatorApplicant.GetAddress().Equals(dstValidator.GetAddress()) {
				continue
			}

			// check not redelegating from the account owned validator
			if srcValidatorApplicant.GetAddress().Equals(accValAddr) {
				continue
			}

			// check if an other redelegation is in progress for the selected account
			if rdInProgress(acc.Address, srcValidatorApplicant.GetAddress(), dstValidator.GetAddress()) {
				continue
			}

			// estimate redelegation amount
			rdAmtDec := sdk.NewDecFromInt(delegation.BondingBalance.Amount)
			if !bondingRD {
				rdAmtDec = sdk.NewDecFromInt(delegation.LPBalance.Amount)
			}
			rdAmt := rdAmtDec.Mul(rdRatio).TruncateInt()
			if rdAmt.IsZero() {
				continue
			}

			targetAcc = acc
			srcValidator = srcValidatorApplicant
			rdCoin = sdk.NewCoin(denom, rdAmt)
			return
		}
	}

	return
}

func redelegateOpHandle(s *Simulator, targetAcc *SimAccount, srcValidator, dstValidator *SimValidator, rdCoin sdk.Coin) {
	s.TxStakeRedelegate(targetAcc, srcValidator.GetAddress(), dstValidator.GetAddress(), rdCoin)
}

func redelegateOpPost(s *Simulator, targetAcc *SimAccount, srcValidator, dstValidator *SimValidator, bondingRD bool) {
	// update validators
	s.UpdateValidator(srcValidator)
	s.UpdateValidator(dstValidator)
	// update account
	s.UpdateAccount(targetAcc)
	// update stats
	if bondingRD {
		s.counter.BRedelegations++
	} else {
		s.counter.LPRedelegations++
	}
}
