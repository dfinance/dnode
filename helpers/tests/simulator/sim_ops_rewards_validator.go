package simulator

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewGetValidatorRewardOp takes validator commissions rewards.
// Op priority:
//   validator - random;
func NewGetValidatorRewardOp(period time.Duration) *SimOperation {
	handler := func(s *Simulator) bool {
		targetAcc, targetVal, rewardCoins := getValidatorRewardOpFindTarget(s)
		if targetAcc == nil || targetVal == nil {
			return false
		}
		getValidatorRewardOpHandle(s, targetAcc, targetVal)

		getValidatorRewardOpPost(s, targetAcc, rewardCoins)
		s.logger.Info(fmt.Sprintf("ValidatorRewardOp: %s for %s: %s", targetVal.GetAddress(), targetAcc.Address, s.FormatCoins(rewardCoins)))

		return true
	}

	return NewSimOperation(period, NewPeriodicNextExecFn(), handler)
}

func getValidatorRewardOpFindTarget(s *Simulator) (targetAcc *SimAccount, targetVal *SimValidator, rewardCoins sdk.Coins) {
	rewardCoins = sdk.NewCoins()

	for _, val := range s.GetAllValidators().GetShuffled() {
		// estimate reward coins
		for _, decCoin := range s.QueryDistValCommission(val.GetAddress()) {
			coin, _ := decCoin.TruncateDecimal()
			rewardCoins = rewardCoins.Add(coin)
		}

		// check there are some rewards
		if rewardCoins.Empty() {
			continue
		}

		targetVal = val
		targetAcc = s.GetAllAccounts().GetByAddress(sdk.AccAddress(targetVal.GetAddress()))
	}

	return
}

func getValidatorRewardOpHandle(s *Simulator, targetAcc *SimAccount, targetVal *SimValidator) {
	s.TxDistValidatorCommission(targetAcc, targetVal.GetAddress())
}

func getValidatorRewardOpPost(s *Simulator, targetAcc *SimAccount, rewardCoins sdk.Coins) {
	// update account
	s.UpdateAccount(targetAcc)
	// update stats
	s.counter.Commissions++
	s.counter.CommissionsCollected = s.counter.CommissionsCollected.Add(rewardCoins...)
}
