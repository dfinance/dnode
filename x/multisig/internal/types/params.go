package types

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/x/params"
)

// Default parameters values.
const (
	// interval in blocks to execute call
	DefIntervalToExecute = 80000
	MinIntervalToExecute = 10
)

// Parameter store key.
var (
	ParamStoreKeyIntervalToExecute = []byte("intervalToExecute")
)

// Params defines genesis params.
type Params struct {
	IntervalToExecute int64 `json:"interval_to_execute" yaml:"interval_to_execute"`
}

// Implements subspace.ParamSet interface.
func (p *Params) ParamSetPairs() params.ParamSetPairs {
	nilValidator := func(value interface{}) error { return nil }

	return params.ParamSetPairs{
		{Key: ParamStoreKeyIntervalToExecute, Value: &p.IntervalToExecute, ValidatorFn: nilValidator},
	}
}

// Equal checks params equality.
func (p *Params) Equal(p2 Params) bool {
	return p.IntervalToExecute == p2.IntervalToExecute
}

// Valid validates params.
func (p Params) Validate() error {
	if p.IntervalToExecute < MinIntervalToExecute {
		return fmt.Errorf("interval to execute calls should be GTE than %d", MinIntervalToExecute)
	}

	return nil
}

func (p Params) String() string {
	return fmt.Sprintf("Params:\n"+
		"IntervalToExecute: %d",
		p.IntervalToExecute,
	)
}

// NewParams creates a new module Params.
func NewParams(intervalToExecute int64) Params {
	return Params{
		IntervalToExecute: intervalToExecute,
	}
}

// ParamKeyTable returns Key declaration for parameters storage.
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&Params{})
}
