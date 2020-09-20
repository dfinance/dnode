package simulator

import (
	"time"

	"github.com/stretchr/testify/require"
)

// NewSimInvariantsOp checks inner simulator state integrity.
func NewSimInvariantsOp(period time.Duration) *SimOperation {
	handler := func(s *Simulator) bool {
		// check validator owner has exactly one self-delegation
		for _, acc := range s.accounts {
			if acc.OperatedValidator != nil {
				selfDelCnt := 0
				for _, del := range acc.Delegations {
					if del.ValidatorAddress.Equals(acc.OperatedValidator.OperatorAddress) {
						selfDelCnt++
					}
				}

				require.Equal(s.t, 1, selfDelCnt, "simInvariants: invalid number of selfDelegations found for: %s", acc.Address)
			}
		}

		// check for duplicated validators
		validatorsMap := make(map[string]bool, len(s.accounts))
		for _, acc := range s.accounts {
			if acc.OperatedValidator != nil {
				valAddrStr := acc.OperatedValidator.OperatorAddress.String()
				found := validatorsMap[valAddrStr]
				require.False(s.t, found, "duplicated validator found: %s", valAddrStr)

				validatorsMap[valAddrStr] = true
			}
		}

		return true
	}

	return NewSimOperation(period, NewPeriodicNextExecFn(), handler)
}

// NewForceUpdateOp updates various simulator states for consistency.
func NewForceUpdateOp(period time.Duration) *SimOperation {
	handler := func(s *Simulator) bool {
		for _, acc := range s.accounts {
			accValidator := acc.OperatedValidator
			if accValidator == nil {
				continue
			}

			s.UpdateValidator(accValidator)
		}

		return true
	}

	return NewSimOperation(period, NewPeriodicNextExecFn(), handler)
}
