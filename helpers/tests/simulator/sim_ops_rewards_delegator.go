package simulator

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type delegatorRewardOpTarget struct {
	Acc *SimAccount
	Val *SimValidator
}

// NewGetDelegatorRewardOp takes all delegators rewards (excluding locked ones).
func NewGetDelegatorRewardOp(period time.Duration) *SimOperation {
	id := "DelegatorRewardOp"

	handler := func(s *Simulator) (bool, string) {
		targets, rewardCoins := getDelegatorRewardOpFindTarget(s)
		if len(targets) == 0 {
			return false, "target not found"
		}
		getDelegatorRewardOpHandle(s, targets)

		getDelegatorRewardOpPost(s, targets, rewardCoins)
		msg := fmt.Sprintf("total from %d targets: %s", len(targets), s.FormatCoins(rewardCoins))

		return true, msg
	}

	return NewSimOperation(id, period, NewPeriodicNextExecFn(), handler)
}

func getDelegatorRewardOpFindTarget(s *Simulator) (targets []delegatorRewardOpTarget, rewardCoins sdk.Coins) {
	rewardCoins = sdk.NewCoins()
	validators := s.GetAllValidators()

	for _, acc := range s.GetAllAccounts() {
		for _, delegation := range acc.GetShuffledDelegations(true) {
			validator := validators.GetByAddress(delegation.ValidatorAddress)
			if validator.RewardsLocked() {
				continue
			}

			// estimate reward coins
			curRewardCoins := sdk.NewCoins()
			for _, decCoin := range s.QueryDistDelReward(acc.Address, delegation.ValidatorAddress) {
				coin, _ := decCoin.TruncateDecimal()
				curRewardCoins = curRewardCoins.Add(coin)
			}

			// check there are some rewards
			if curRewardCoins.Empty() {
				continue
			}

			targets = append(targets, delegatorRewardOpTarget{
				Acc: acc,
				Val: validator,
			})
			rewardCoins = rewardCoins.Add(curRewardCoins...)
		}
	}

	return
}

func getDelegatorRewardOpHandle(s *Simulator, targets []delegatorRewardOpTarget) {
	for _, target := range targets {
		s.TxDistDelegatorRewards(target.Acc, target.Val.GetAddress())
	}
}

func getDelegatorRewardOpPost(s *Simulator, targets []delegatorRewardOpTarget, rewardCoins sdk.Coins) {
	// update accounts
	for _, target := range targets {
		s.UpdateAccount(target.Acc)
	}
	// update stats
	s.counter.RewardsWithdraws += int64(len(targets))
	s.counter.RewardsCollectedMain = s.counter.RewardsCollectedMain.Add(rewardCoins.AmountOf(s.mainDenom))
	s.counter.RewardsCollectedStaking = s.counter.RewardsCollectedStaking.Add(rewardCoins.AmountOf(s.stakingDenom))
}
