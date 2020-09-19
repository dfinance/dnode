package simulator

import (
	"time"

	"github.com/stretchr/testify/require"
)

// NewSimInvariantsOp checks inner simulator state integrity.
func NewSimInvariantsOp(period time.Duration) *SimOperation {
	handler := func(s *Simulator) bool {
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

		return true
	}

	return NewSimOperation(period, NewPeriodicNextExecFn(), handler)
}
