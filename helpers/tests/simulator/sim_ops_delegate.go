package simulator

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

// NewDelegateBondingOp picks a validator and searches for an account to delegate bonding tokens.
// SelfStake increment is allowed.
// Delegation amount = current account balance * {delegateRatio}.
// Delegation is allowed if ratio (current staking bonding pools supply / total bonding tokens supply) < {maxBondingRatio}.
// Op priorities:
//   validator:
//     - bonded;
//     - lowest bonding tokens amount;
//   account:
//     - highest bonding tokens balance;
//     - enough coins;
func NewDelegateBondingOp(period time.Duration, delegateRatio, maxBondingRatio sdk.Dec) *SimOperation {
	checkRatioArg("DelegateBondingOp", "delegateRatio", delegateRatio)
	checkRatioArg("DelegateBondingOp", "maxBondingRatio", maxBondingRatio)

	handler := func(s *Simulator) bool {
		if delegateOpCheckInput(s, true, maxBondingRatio) {
			return true
		}

		targetVal, targetAcc, delCoin := delegateOpFindTarget(s, true, delegateRatio)
		if targetVal == nil || targetAcc == nil {
			return false
		}

		if delegateOpHandle(s, targetVal, targetAcc, delCoin) {
			s.logger.Error(fmt.Sprintf("DelegateBondingOp: %s: overflow", targetVal.GetAddress()))
			return false
		}

		delegateOpPost(s, targetVal, targetAcc, true)
		s.logger.Info(fmt.Sprintf("DelegateBondingOp: %s: %s -> %s", targetAcc.Address, s.FormatCoin(delCoin), targetVal.GetAddress()))

		return true
	}

	return NewSimOperation(period, NewPeriodicNextExecFn(), handler)
}

// NewDelegateLPOp picks a validator and searches for an account to delegate LP tokens.
// Delegation amount = current account balance * {delegateRatio}.
// Delegation is allowed if ratio (current staking LP pool supply / total LP tokens supply) < {maxBondingRatio}.
// Op priorities:
//   validator:
//     - bonded;
//     - lowest LP tokens amount;
//   account:
//     - highest LP tokens balance;
//     - enough coins;
func NewDelegateLPOp(period time.Duration, delegateRatio, maxBondingRatio sdk.Dec) *SimOperation {
	checkRatioArg("DelegateLPOp", "delegateRatio", delegateRatio)
	checkRatioArg("DelegateLPOp", "maxBondingRatio", maxBondingRatio)

	handler := func(s *Simulator) bool {
		if delegateOpCheckInput(s, false, maxBondingRatio) {
			return true
		}

		targetVal, targetAcc, delCoin := delegateOpFindTarget(s, false, delegateRatio)
		if targetVal == nil || targetAcc == nil {
			return false
		}

		if overflow := delegateOpHandle(s, targetVal, targetAcc, delCoin); overflow {
			require.False(s.t, overflow)
		}

		delegateOpPost(s, targetVal, targetAcc, false)
		s.logger.Info(fmt.Sprintf("DelegateLPOp: %s: %s -> %s", targetAcc.Address, s.FormatCoin(delCoin), targetVal.GetAddress()))

		return true
	}

	return NewSimOperation(period, NewPeriodicNextExecFn(), handler)
}

func delegateOpCheckInput(s *Simulator, bondingD bool, maxRatio sdk.Dec) (stop bool) {
	pool := s.QueryStakePools()

	var stakingSupply, totalSupply sdk.Int
	if bondingD {
		stakingSupply = pool.BondedTokens.Add(pool.NotBondedTokens)
		totalSupply = s.QuerySupplyTotal().AmountOf(s.stakingDenom)
	} else {
		stakingSupply = pool.LiquidityTokens
		totalSupply = s.QuerySupplyTotal().AmountOf(s.lpDenom)
	}

	// check staking pool total supply to all tokens supply ratio
	curRatio := stakingSupply.ToDec().Quo(totalSupply.ToDec())
	if curRatio.GT(maxRatio) {
		return true
	}

	return false
}

func delegateOpFindTarget(s *Simulator, bondingD bool, delegateRatio sdk.Dec) (targetVal *SimValidator, targetAcc *SimAccount, delCoin sdk.Coin) {
	denom := s.stakingDenom
	if !bondingD {
		denom = s.lpDenom
	}

	// pick a bonded validator with the lowest tokens amount
	validators := s.GetValidators(true, false, false).GetSortedByTokens(bondingD, false)
	if len(validators) == 0 {
		return
	}
	targetVal = validators[0]

	// pick an account with max tokens
	var delAmt sdk.Int
	for _, acc := range s.GetAllAccounts().GetSortedByBalance(denom, true) {
		// estimate delegation amount
		accCoinAmtDec := sdk.NewDecFromInt(acc.Coins.AmountOf(denom))
		delAmt = accCoinAmtDec.Mul(delegateRatio).TruncateInt()
		if delAmt.IsZero() {
			continue
		}

		targetAcc = acc
		delCoin = sdk.NewCoin(denom, delAmt)
	}

	return
}

func delegateOpHandle(s *Simulator, targetVal *SimValidator, targetAcc *SimAccount, delCoin sdk.Coin) (stop bool) {
	overflow := s.TxStakeDelegate(targetAcc, targetVal, delCoin)
	if overflow {
		stop = true
	}

	return
}

func delegateOpPost(s *Simulator, targetVal *SimValidator, targetAcc *SimAccount, bondingD bool) {
	// update account
	s.UpdateAccount(targetAcc)
	// update validator
	s.UpdateValidator(targetVal)
	// update stats
	if bondingD {
		s.counter.BDelegations++
	} else {
		s.counter.LPDelegations++
	}
}
