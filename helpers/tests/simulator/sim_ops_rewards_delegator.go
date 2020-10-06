package simulator

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewGetDelegatorRewardOp takes delegator rewards.
// Op priority:
//   account;
//     - random;
//     - has delegations;
//   validator
//     - random account delegation;
//     - rewards are not locked;
func NewGetDelegatorRewardOp(period time.Duration) *SimOperation {
	id := "DelegatorRewardOp"

	handler := func(s *Simulator) (bool, string) {
		targetAcc, targetVal, rewardCoins := getDelegatorRewardOpFindTarget(s)
		if targetAcc == nil || targetVal == nil {
			return false, "target not found"
		}
		getDelegatorRewardOpHandle(s, targetAcc, targetVal)

		getDelegatorRewardOpPost(s, targetAcc, rewardCoins)
		msg := fmt.Sprintf("%s from %s: %s", targetAcc.Address, targetVal.GetAddress(), s.FormatCoins(rewardCoins))

		return true, msg
	}

	return NewSimOperation(id, period, NewPeriodicNextExecFn(), handler)
}

func getDelegatorRewardOpFindTarget(s *Simulator) (targetAcc *SimAccount, targetVal *SimValidator, rewardCoins sdk.Coins) {
	rewardCoins = sdk.NewCoins()
	validators := s.GetAllValidators()

	for _, acc := range s.GetAllAccounts().GetShuffled() {
		for _, delegation := range acc.GetShuffledDelegations(true) {
			validator := validators.GetByAddress(delegation.ValidatorAddress)
			if validator.RewardsLocked() {
				continue
			}

			// estimate reward coins
			for _, decCoin := range s.QueryDistDelReward(acc.Address, delegation.ValidatorAddress) {
				coin, _ := decCoin.TruncateDecimal()
				rewardCoins = rewardCoins.Add(coin)
			}

			// check there are some rewards
			if rewardCoins.Empty() {
				continue
			}

			targetAcc = acc
			targetVal = validator
			return
		}
	}

	return
}

func getDelegatorRewardOpHandle(s *Simulator, targetAcc *SimAccount, targetVal *SimValidator) {
	s.TxDistDelegatorRewards(targetAcc, targetVal.GetAddress())
}

func getDelegatorRewardOpPost(s *Simulator, targetAcc *SimAccount, rewardCoins sdk.Coins) {
	// update account
	s.UpdateAccount(targetAcc)
	// update stats
	s.counter.Rewards++
	s.counter.RewardsCollected = s.counter.RewardsCollected.Add(rewardCoins...)
}
