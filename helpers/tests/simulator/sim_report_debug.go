package simulator

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

type SimDebugReportItem struct {
	Validators []DebugValidatorData
	Accounts   []DebugAccoutData
}

type DebugValidatorData struct {
	Validator   staking.Validator
	Delegations staking.DelegationResponses
}

type DebugAccoutData struct {
	Address            sdk.AccAddress
	MainCoinBalance    sdk.Int
	StakingCoinBalance sdk.Int
}

func (r SimDebugReportItem) String() string {
	str := strings.Builder{}

	str.WriteString("Validators:\n")
	for i, valData := range r.Validators {
		var selfDelegation *staking.DelegationResponse
		for i := 0; i < len(valData.Delegations); i++ {
			del := &valData.Delegations[i]
			if sdk.ValAddress(del.DelegatorAddress).Equals(valData.Validator.OperatorAddress) {
				selfDelegation = del
				break
			}
		}

		str.WriteString(fmt.Sprintf("[%03d] %s\n", i, valData.Validator.OperatorAddress))
		str.WriteString(fmt.Sprintf("  Address:    %s\n", valData.Validator.OperatorAddress))
		str.WriteString(fmt.Sprintf("  Status:     %s\n", valData.Validator.Status))
		str.WriteString(fmt.Sprintf("  Jailed:     %v\n", valData.Validator.Jailed))
		str.WriteString(fmt.Sprintf("  BTokens:    %s\n", valData.Validator.Bonding.Tokens))
		str.WriteString(fmt.Sprintf("  BDelShares: %s\n", valData.Validator.Bonding.DelegatorShares))
		str.WriteString(fmt.Sprintf("  MinSDel:    %s\n", valData.Validator.MinSelfDelegation))
		str.WriteString(fmt.Sprintf("  UBTime:     %s\n", valData.Validator.UnbondingCompletionTime))
		//
		str.WriteString(fmt.Sprintf("    Dels: count:   %d\n", len(valData.Delegations)))
		if selfDelegation != nil {
			str.WriteString(fmt.Sprintf("    Dels: SelfAmt: %s\n", selfDelegation.BondingBalance.Amount))
			str.WriteString(fmt.Sprintf("    Dels: SelfShr: %s\n", selfDelegation.BondingShares))
			str.WriteString(fmt.Sprintf("    Dels: SelfAdr: %s\n", selfDelegation.DelegatorAddress))
		} else {
			str.WriteString("    Dels: SelfAmt: nil\n")
		}
	}

	str.WriteString("Accounts:\n")
	for i, accData := range r.Accounts {
		str.WriteString(fmt.Sprintf("[%03d] %s\n", i, accData.Address))
		str.WriteString(fmt.Sprintf("  Balance (main):    %s\n", accData.MainCoinBalance))
		str.WriteString(fmt.Sprintf("  Balance (staking): %s\n", accData.StakingCoinBalance))
	}

	return str.String()
}

func BuildDebugReportItem(s *Simulator) SimDebugReportItem {
	r := SimDebugReportItem{}

	for _, v := range s.QueryReadAllValidators() {
		dels := s.QueryStakeValDelegations(&v)
		r.Validators = append(r.Validators, DebugValidatorData{
			Validator:   v,
			Delegations: dels,
		})
	}

	for _, acc := range s.GetAccountsSortedByBalance(true) {
		r.Accounts = append(r.Accounts, DebugAccoutData{
			Address:            acc.Address,
			MainCoinBalance:    acc.Coins.AmountOf(s.mainDenom),
			StakingCoinBalance: acc.Coins.AmountOf(s.stakingDenom),
		})
	}

	return r
}
