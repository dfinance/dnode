package simulator

import (
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type validatorRewardOp struct {
	Acc *SimAccount
	Val *SimValidator
}

// NewGetValidatorRewardOp takes all validators commissions rewards.
func NewGetValidatorRewardOp(period time.Duration) *SimOperation {
	id := "ValidatorRewardOp"

	handler := func(s *Simulator) (bool, string) {
		targets, rewardCoins := getValidatorRewardOpFindTarget(s)
		if len(targets) == 0 {
			return false, "target not found"
		}

		if stopMsg := getValidatorRewardOpHandle(s, targets); stopMsg != "" {
			msg := fmt.Sprintf("withdraw validator commission failed: %s", stopMsg)
			return false, msg
		}

		getValidatorRewardOpPost(s, targets, rewardCoins)
		msg := fmt.Sprintf("total from %d targets: %s", len(targets), s.FormatCoins(rewardCoins))

		return true, msg
	}

	return NewSimOperation(id, period, NewPeriodicNextExecFn(), handler)
}

func getValidatorRewardOpFindTarget(s *Simulator) (targets []validatorRewardOp, rewardCoins sdk.Coins) {
	rewardCoins = sdk.NewCoins()

	for _, val := range s.GetAllValidators().GetShuffled() {
		// check there are some commission rewards available
		decCoins := s.QueryDistValCommission(val.GetAddress())
		if decCoins.Empty() {
			continue
		}

		// estimate reward coins
		curRewardCoins := sdk.NewCoins()
		for _, decCoin := range decCoins {
			coin, _ := decCoin.TruncateDecimal()
			curRewardCoins = curRewardCoins.Add(coin)
		}

		targets = append(targets, validatorRewardOp{
			Acc: s.GetAllAccounts().GetByAddress(val.GetOperatorAddress()),
			Val: val,
		})
		rewardCoins = rewardCoins.Add(curRewardCoins...)
	}

	return
}

func getValidatorRewardOpHandle(s *Simulator, targets []validatorRewardOp) (stopMsg string) {
	for _, target := range targets {
		if s.TxDistValidatorCommission(target.Acc, target.Val.GetAddress()) {
			stopMsg = fmt.Sprintf("targetVal %s", target.Val.GetAddress())
			return
		}
	}

	return
}

func getValidatorRewardOpPost(s *Simulator, targets []validatorRewardOp, rewardCoins sdk.Coins) {
	// update account
	for _, target := range targets {
		s.UpdateAccount(target.Acc)
	}
	// update stats
	s.counter.CommissionWithdraws += int64(len(targets))
	s.counter.CommissionsCollectedMain = s.counter.CommissionsCollectedMain.Add(rewardCoins.AmountOf(s.mainDenom))
	s.counter.CommissionsCollectedStaking = s.counter.CommissionsCollectedStaking.Add(rewardCoins.AmountOf(s.stakingDenom))
}
