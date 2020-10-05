package simulator

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewGetValidatorRewardOp takes validator commissions rewards.
// Op priority:
//   validator - random;
func NewLockValidatorRewardsOp(period time.Duration, maxLockedRatio sdk.Dec) *SimOperation {
	handler := func(s *Simulator) bool {
		if lockValidatorRewardsOpCheckInput(s, maxLockedRatio) {
			return true
		}

		targetAcc, targetVal := lockValidatorRewardsOpFindTarget(s)
		if targetAcc == nil || targetVal == nil {
			return false
		}
		lockValidatorRewardsOpHandle(s, targetAcc, targetVal)

		lockValidatorRewardsOpPost(s, targetVal)
		s.logger.Info(fmt.Sprintf("LockValidatorRewardsOp: %s for %s", targetVal.GetAddress(), targetAcc.Address))

		return true
	}

	return NewSimOperation(period, NewPeriodicNextExecFn(), handler)
}

func lockValidatorRewardsOpCheckInput(s *Simulator, maxRatio sdk.Dec) (stop bool) {
	// check current locked ratio
	validators := s.GetValidators(true, false, false)
	locked := len(validators.GetLocked())
	total := len(validators)

	curRatio := sdk.NewDec(int64(locked)).Quo(sdk.NewDec(int64(total)))
	if curRatio.GT(maxRatio) {
		stop = true
		return
	}

	return
}

func lockValidatorRewardsOpFindTarget(s *Simulator) (targetAcc *SimAccount, targetVal *SimValidator) {
	for _, val := range s.GetAllValidators().GetShuffled() {
		if val.RewardsLocked() {
			continue
		}

		targetAcc = s.GetAllAccounts().GetByAddress(val.GetOperatorAddress())
		targetVal = val
		break
	}

	return
}

func lockValidatorRewardsOpHandle(s *Simulator, targetAcc *SimAccount, targetVal *SimValidator) {
	// lock and disable auto-renewal
	s.TxDistLockRewards(targetAcc, targetVal.GetAddress())
	s.TxDistDisableAutoRenewal(targetAcc, targetVal.GetAddress())
}

func lockValidatorRewardsOpPost(s *Simulator, targetVal *SimValidator) {
	// update validator
	s.UpdateValidator(targetVal)
	// update stats
	s.counter.LockedRewards++
}
