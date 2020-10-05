package simulator

import (
	"time"

	"github.com/stretchr/testify/require"
)

// NewSimInvariantsOp checks inner simulator state integrity.
func NewSimInvariantsOp(period time.Duration) *SimOperation {
	handler := func(s *Simulator) bool {
		// check validator owner has exactly one self-delegation
		for _, acc := range s.GetAllAccounts() {
			if acc.IsValOperator() {
				selfDelCnt := 0
				for _, del := range acc.Delegations {
					if del.ValidatorAddress.Equals(acc.OperatedValidator.GetAddress()) {
						selfDelCnt++
					}
				}

				require.Equal(s.t, 1, selfDelCnt, "simInvariants: invalid number of selfDelegations found for: %s", acc.Address)
			}
		}

		// check for duplicated validators
		validatorsMap := make(map[string]bool, len(s.accounts))
		for _, val := range s.GetAllValidators() {
			valAddrStr := val.GetAddress().String()
			found := validatorsMap[valAddrStr]
			require.False(s.t, found, "duplicated validator found: %s", valAddrStr)

			validatorsMap[valAddrStr] = true
		}

		return true
	}

	return NewSimOperation(period, NewPeriodicNextExecFn(), handler)
}

// NewForceUpdateOp updates various simulator states for consistency.
func NewForceUpdateOp(period time.Duration) *SimOperation {
	handler := func(s *Simulator) bool {
		for _, val := range s.GetAllValidators() {
			s.UpdateValidator(val)
		}

		s.counter.LockedRewards = int64(len(s.GetAllValidators().GetLocked()))

		return true
	}

	return NewSimOperation(period, NewPeriodicNextExecFn(), handler)
}
