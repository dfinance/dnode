package simulator

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/stretchr/testify/require"
	"math/rand"
	"sort"
)

func (s *Simulator) GetRandomAccount() *SimAccount {
	aMaxIndex := len(s.accounts) - 1
	return s.accounts[rand.Intn(aMaxIndex)]
}

// UpdateAccount updates account balance.
func (s *Simulator) UpdateAccount(simAcc *SimAccount) {
	require.NotNil(s.t, simAcc)

	updAcc := s.QueryAuthAccount(simAcc.Address)
	simAcc.Coins = updAcc.GetCoins()
}

// GetValidatorByAddress returns validator.
func (s *Simulator) GetValidatorByAddress(address sdk.ValAddress) *staking.Validator {
	for _, acc := range s.accounts {
		if acc.OperatedValidator != nil {
			if acc.OperatedValidator.OperatorAddress.Equals(address) {
				return acc.OperatedValidator
			}
		}
	}

	return nil
}

// UpdateValidator updates validator status.
func (s *Simulator) UpdateValidator(val *staking.Validator) {
	require.NotNil(s.t, val)

	updVal := s.QueryStakingValidator(val.OperatorAddress)
	val.Status = updVal.Status
	val.Jailed = updVal.Jailed
	val.Tokens = updVal.Tokens
	val.DelegatorShares = updVal.DelegatorShares
	val.UnbondingHeight = updVal.UnbondingHeight
	val.UnbondingCompletionTime = updVal.UnbondingCompletionTime
}

// GetValidators returns all known to Simulator validators.
func (s *Simulator) GetValidators() []*staking.Validator {
	validators := make([]*staking.Validator, 0)
	for _, acc := range s.accounts {
		if acc.OperatedValidator != nil {
			validators = append(validators, acc.OperatedValidator)
		}
	}

	return validators
}

// GetValidatorWithMinimalStake returns validator with minimal stake or false in second value if validator not found.
func (s *Simulator) GetValidatorSortedByStake(desc bool) []*staking.Validator {
	validators := s.GetValidators()

	sort.Slice(validators, func(i, j int) bool {
		if validators[i].Tokens.GT(validators[j].Tokens) {
			return desc
		}
		return !desc
	})

	return validators
}

// GetShuffledAccounts returns random sorted account list.
func (s Simulator) GetShuffledAccounts() []*SimAccount {
	tmpAcc := make([]*SimAccount, len(s.accounts))
	copy(tmpAcc, s.accounts)

	for i := range tmpAcc {
		j := rand.Intn(i + 1)
		tmpAcc[i], tmpAcc[j] = tmpAcc[j], tmpAcc[i]
	}

	return tmpAcc
}

// GetSortedDelegation returns delegation sorted list.
func GetSortedDelegation(responses staking.DelegationResponses, desc bool) staking.DelegationResponses {
	sort.Slice(responses, func(i, j int) bool {
		if responses[i].Balance.Amount.GT(responses[j].Balance.Amount) {
			return desc
		}
		return !desc
	})

	return responses
}

// ShuffleRewards returns rewards in the random order.
func ShuffleRewards(rewards []distribution.DelegationDelegatorReward) []distribution.DelegationDelegatorReward {
	tmp := make([]distribution.DelegationDelegatorReward, len(rewards))
	copy(tmp, rewards)

	for i := range tmp {
		j := rand.Intn(i + 1)
		tmp[i], tmp[j] = tmp[j], tmp[i]
	}

	return tmp
}
