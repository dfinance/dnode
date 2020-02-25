package types

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/x/params"
)

// Default parameters.
const (
	DefIntervalToExecute = 86400 // interval in blocks to execute proposal.
	MinIntervalToExecute = 10
)

// Keys to store parameters.
var (
	KeyIntervalToExecute = []byte("IntervalToExecute")
)

// Describing parameters for multisig module.
type Params struct {
	IntervalToExecute int64 `json:"interval_to_execute"`
}

// Create new instance to store parameters.
func NewParams(intervalToExecute int64) Params {
	return Params{
		IntervalToExecute: intervalToExecute,
	}
}

func (p *Params) ParamSetPairs() params.ParamSetPairs {
	return params.ParamSetPairs{
		{Key: KeyIntervalToExecute, Value: &p.IntervalToExecute},
	}
}

func (p *Params) Equal(p2 Params) bool {
	return p.IntervalToExecute == p2.IntervalToExecute
}

func (p Params) Validate() error {
	if p.IntervalToExecute < MinIntervalToExecute {
		return fmt.Errorf("interval to execute calls should be not less %d", MinIntervalToExecute)
	}

	return nil
}

func (p Params) String() string {
	return fmt.Sprintf("\tIntervalToExecute: %d", p.IntervalToExecute)
}

func DefaultParams() Params {
	return NewParams(DefIntervalToExecute)
}
