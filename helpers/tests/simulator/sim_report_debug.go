package simulator

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

type SimDebugReportItem struct {
	Validators []DebugValidatorData
}

type DebugValidatorData struct {
	Validator   staking.Validator
	Delegations staking.DelegationResponses
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
		str.WriteString(fmt.Sprintf("  Address:   %s\n", valData.Validator.OperatorAddress))
		str.WriteString(fmt.Sprintf("  Status:    %s\n", valData.Validator.Status))
		str.WriteString(fmt.Sprintf("  Jailed:    %v\n", valData.Validator.Jailed))
		str.WriteString(fmt.Sprintf("  Tokens:    %s\n", valData.Validator.Tokens))
		str.WriteString(fmt.Sprintf("  MinSDel:   %s\n", valData.Validator.MinSelfDelegation))
		str.WriteString(fmt.Sprintf("  DelShares: %s\n", valData.Validator.DelegatorShares))
		str.WriteString(fmt.Sprintf("  UBTime:    %s\n", valData.Validator.UnbondingCompletionTime))
		//
		str.WriteString(fmt.Sprintf("    Dels: count:   %d\n", len(valData.Delegations)))
		if selfDelegation != nil {
			str.WriteString(fmt.Sprintf("    Dels: SelfAmt: %s\n", selfDelegation.Balance.Amount))
			str.WriteString(fmt.Sprintf("    Dels: SelfShr: %s\n", selfDelegation.Shares))
			str.WriteString(fmt.Sprintf("    Dels: SelfAdr: %s\n", selfDelegation.DelegatorAddress))
		} else {
			str.WriteString("    Dels: SelfAmt: nil\n")
		}
	}

	return str.String()
}

func BuildDebugReportItem(s *Simulator) SimDebugReportItem {
	r := SimDebugReportItem{}

	// get all validators with delegations
	valsPage := 1
	validators := make([]staking.Validator, 0)
	for {
		rcvBondedVals := s.QueryStakeValidators(valsPage, 100, sdk.Bonded.String())
		rcvUnbondingVals := s.QueryStakeValidators(valsPage, 100, sdk.Unbonding.String())
		rcvUnbondedVals := s.QueryStakeValidators(valsPage, 100, sdk.Unbonded.String())

		validators = append(validators, rcvBondedVals...)
		validators = append(validators, rcvUnbondingVals...)
		validators = append(validators, rcvUnbondedVals...)

		if (len(rcvBondedVals) + len(rcvUnbondingVals) + len(rcvUnbondedVals)) == 0 {
			break
		}
		valsPage++
	}

	for _, v := range validators {
		dels := s.QueryStakeValDelegations(&v)
		r.Validators = append(r.Validators, DebugValidatorData{
			Validator:   v,
			Delegations: dels,
		})
	}

	return r
}
