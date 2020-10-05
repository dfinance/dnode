package simulator

import (
	"math/rand"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/stretchr/testify/require"
)

func (s *Simulator) GetRandomAccount() *SimAccount {
	aMaxIndex := len(s.accounts) - 1
	return s.accounts[rand.Intn(aMaxIndex)]
}

// UpdateAccount updates account balance and active delegations.
func (s *Simulator) UpdateAccount(simAcc *SimAccount) {
	require.NotNil(s.t, simAcc)

	updAcc := s.QueryAuthAccount(simAcc.Address)
	simAcc.Coins = updAcc.GetCoins()
	simAcc.Delegations = s.QueryStakeDelDelegations(simAcc.Address)
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

	updVal := s.QueryStakeValidator(val.OperatorAddress)
	val.Status = updVal.Status
	val.Jailed = updVal.Jailed
	val.Bonding = updVal.Bonding
	val.LP = updVal.LP
	val.UnbondingHeight = updVal.UnbondingHeight
	val.UnbondingCompletionTime = updVal.UnbondingCompletionTime
}

// GetValidators returns all known to Simulator validators.
func (s *Simulator) GetValidators(bonded, unbonding, unbonded bool) []*staking.Validator {
	validators := make([]*staking.Validator, 0)
	for _, acc := range s.accounts {
		if acc.OperatedValidator != nil {
			add := false
			switch acc.OperatedValidator.Status {
			case sdk.Bonded:
				if bonded {
					add = true
				}
			case sdk.Unbonding:
				if unbonding {
					add = true
				}
			case sdk.Unbonded:
				if unbonded {
					add = true
				}
			}

			if add {
				validators = append(validators, acc.OperatedValidator)
			}
		}
	}

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

// GetAccountsSortedByBalance returns account sorted by staking denom list.
func (s Simulator) GetAccountsSortedByBalance(desc bool) []*SimAccount {
	tmpAccs := make([]*SimAccount, len(s.accounts))
	copy(tmpAccs, s.accounts)

	sort.Slice(tmpAccs, func(i, j int) bool {
		iBalance := tmpAccs[i].Coins.AmountOf(s.stakingDenom)
		jBalance := tmpAccs[j].Coins.AmountOf(s.stakingDenom)

		if iBalance.GT(jBalance) {
			return desc
		}
		return !desc
	})

	return tmpAccs
}

func (s *Simulator) FormatStakingCoin(coin sdk.Coin) string {
	return s.FormatIntDecimals(coin.Amount, s.stakingAmountDecimalsRatio) + s.stakingDenom
}

func (s *Simulator) FormatIntDecimals(value sdk.Int, decRatio sdk.Dec) string {
	valueDec := sdk.NewDecFromInt(value)
	fixedDec := valueDec.Mul(decRatio)

	return fixedDec.String()
}

func (s *Simulator) FormatDecDecimals(value sdk.Dec, decRatio sdk.Dec) string {
	fixedDec := value.Mul(decRatio)

	return fixedDec.String()
}

// GetSortedByStakeValidator returns validators sorted by stake.
func GetSortedByStakeValidator(validators []*staking.Validator, desc bool) []*staking.Validator {
	sort.Slice(validators, func(i, j int) bool {
		if validators[i].Bonding.Tokens.GT(validators[j].Bonding.Tokens) {
			return desc
		}
		return !desc
	})

	return validators
}

// GetSortedDelegation returns delegation sorted list.
func GetSortedDelegation(responses staking.DelegationResponses, desc bool) staking.DelegationResponses {
	sort.Slice(responses, func(i, j int) bool {
		if responses[i].BondingBalance.Amount.GT(responses[j].BondingBalance.Amount) {
			return desc
		}
		return !desc
	})

	return responses
}

// GetShuffledDelegations returns delegations in the random order.
func GetShuffledDelegations(delegations staking.DelegationResponses) staking.DelegationResponses {
	tmp := make(staking.DelegationResponses, len(delegations))
	copy(tmp, delegations)

	for i := range tmp {
		j := rand.Intn(i + 1)
		tmp[i], tmp[j] = tmp[j], tmp[i]
	}

	return tmp
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
