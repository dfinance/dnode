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
	id := "ValidatorRewardOp"

	handler := func(s *Simulator) (bool, string) {
		targetAcc, targetVal, rewardCoins := getValidatorRewardOpFindTarget(s)
		if targetAcc == nil || targetVal == nil {
			return false, "target not found"
		}

		if getValidatorRewardOpHandle(s, targetAcc, targetVal) {
			msg := fmt.Sprintf("can't withdraw %s validator commission", targetVal.GetAddress())
			return false, msg
		}

		getValidatorRewardOpPost(s, targetAcc, rewardCoins)
		msg := fmt.Sprintf("%s for %s: %s", targetVal.GetAddress(), targetAcc.Address, s.FormatCoins(rewardCoins))

		return true, msg
	}

	return NewSimOperation(id, period, NewPeriodicNextExecFn(), handler)
}

func getValidatorRewardOpFindTarget(s *Simulator) (targetAcc *SimAccount, targetVal *SimValidator, rewardCoins sdk.Coins) {
	rewardCoins = sdk.NewCoins()

	for _, val := range s.GetAllValidators().GetShuffled() {
		// check there are some commission rewards available
		decCoins := s.QueryDistValCommission(val.GetAddress())
		if decCoins.Empty() {
			continue
		}

		// estimate reward coins
		for _, decCoin := range decCoins {
			coin, _ := decCoin.TruncateDecimal()
			rewardCoins = rewardCoins.Add(coin)
		}

		targetVal = val
		targetAcc = s.GetAllAccounts().GetByAddress(targetVal.GetOperatorAddress())
	}

	return
}

func getValidatorRewardOpHandle(s *Simulator, targetAcc *SimAccount, targetVal *SimValidator) (stop bool) {
	if s.TxDistValidatorCommission(targetAcc, targetVal.GetAddress()) {
		stop = true
	}

	return
}

func getValidatorRewardOpPost(s *Simulator, targetAcc *SimAccount, rewardCoins sdk.Coins) {
	// update account
	s.UpdateAccount(targetAcc)
	// update stats
	s.counter.Commissions++
	s.counter.CommissionsCollected = s.counter.CommissionsCollected.Add(rewardCoins...)
}
