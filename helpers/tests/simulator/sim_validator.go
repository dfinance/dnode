package simulator

import (
	"math/rand"
	"sort"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/stretchr/testify/require"
)

type SimValidatorConfig struct {
	Commission staking.CommissionRates
}

type SimValidator struct {
	Validator          staking.Validator
	RewardsLockedUntil time.Time
}

// GetAddress returns validator address.
func (v *SimValidator) GetAddress() sdk.ValAddress {
	return v.Validator.OperatorAddress
}

// GetOperatorAddress returns validator operator address.
func (v *SimValidator) GetOperatorAddress() sdk.AccAddress {
	return sdk.AccAddress(v.Validator.OperatorAddress)
}

// GetStatus returns validator bonding status.
func (v *SimValidator) GetStatus() sdk.BondStatus {
	return v.Validator.Status
}

// RewardsLocked check if validator rewards are locked.
func (v *SimValidator) RewardsLocked() bool {
	return !v.RewardsLockedUntil.IsZero()
}

// LockRewards updates locked rewards state.
func (v *SimValidator) LockRewards(until time.Time) {
	v.RewardsLockedUntil = until
}

// UnlockRewards updates locked rewards state.
func (v *SimValidator) UnlockRewards() {
	v.RewardsLockedUntil = time.Time{}
}

// NewSimValidator createes a new SimValidator object.
func NewSimValidator(val staking.Validator) *SimValidator {
	return &SimValidator{
		Validator:          val,
		RewardsLockedUntil: time.Time{},
	}
}

type SimValidators []*SimValidator

// GetByAddress returns validator by address.
func (v SimValidators) GetByAddress(address sdk.ValAddress) *SimValidator {
	for _, val := range v {
		if val.GetAddress().Equals(address) {
			return val
		}
	}

	return nil
}

// GetShuffled returns random sorted validators list.
func (v SimValidators) GetShuffled() SimValidators {
	tmpVal := make(SimValidators, len(v))
	copy(tmpVal, v)

	for i := range tmpVal {
		j := rand.Intn(i + 1)
		tmpVal[i], tmpVal[j] = tmpVal[j], tmpVal[i]
	}

	return tmpVal
}

// GetSortedByTokens returns validators list sorted by tokens amount.
func (v SimValidators) GetSortedByTokens(bondingTokens, desc bool) SimValidators {
	tmpVals := make(SimValidators, len(v))
	copy(tmpVals, v)

	sort.Slice(tmpVals, func(i, j int) bool {
		if bondingTokens {
			if tmpVals[i].Validator.GetBondingTokens().GT(tmpVals[j].Validator.GetBondingTokens()) {
				return desc
			}
			return !desc
		}

		if tmpVals[i].Validator.GetLPTokens().GT(tmpVals[j].Validator.GetLPTokens()) {
			return desc
		}
		return !desc
	})

	return tmpVals
}

// GetLocked returns validators list with locked rewards.
func (v SimValidators) GetLocked() SimValidators {
	tmpVals := make(SimValidators, 0, len(v))
	for _, val := range v {
		if val.RewardsLocked() {
			tmpVals = append(tmpVals, val)
		}
	}

	return tmpVals
}

// UpdateValidator updates validator status.
func (s *Simulator) UpdateValidator(val *SimValidator) {
	require.NotNil(s.t, val)

	updVal := s.QueryStakeValidator(val.GetAddress())
	val.Validator.Status = updVal.Status
	val.Validator.Jailed = updVal.Jailed
	val.Validator.Bonding = updVal.Bonding
	val.Validator.LP = updVal.LP
	val.Validator.UnbondingHeight = updVal.UnbondingHeight
	val.Validator.UnbondingCompletionTime = updVal.UnbondingCompletionTime

	updLState := s.QueryDistLockState(val.GetAddress())
	if updLState.Enabled {
		val.LockRewards(updLState.UnlocksAt)
	} else {
		val.UnlockRewards()
	}
}
