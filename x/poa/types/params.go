package types

import (
	"github.com/cosmos/cosmos-sdk/x/params"
	"fmt"
)

const (
	DefaultMaxValidators uint16 = 11
	DefaultMinValidators uint16 = 3
)

var (
	KeyMaxValidators = []byte("max_validators")
	KeyMinValidators = []byte("min_validators")
)

type Params struct {
	MaxValidators uint16
	MinValidtors  uint16
}

func NewParams(maxValidators uint16, minValidators uint16) Params {
	return Params{
		MaxValidators: maxValidators,
		MinValidtors:  minValidators,
	}
}

func (p* Params) ParamSetPairs() params.ParamSetPairs {
	return params.ParamSetPairs{
		{KeyMaxValidators, &p.MaxValidators},
		{KeyMinValidators, &p.MinValidtors},
	}
}

func (p* Params) Equal(p2 Params) bool {
	return p.MinValidtors == p2.MinValidtors &&
		p.MaxValidators == p2.MaxValidators
}

func (p Params) Validate() error {
	if p.MinValidtors < DefaultMinValidators {
		return fmt.Errorf("minimum amount of validators should be not less %d", DefaultMinValidators)
	}

	return nil
}

func (p Params) String() string {
	return fmt.Sprintf("\tMax Validators: %d\tMin Validators: %d", p.MaxValidators, p.MaxValidators)
}

func DefaultParams() Params {
	return NewParams(DefaultMaxValidators, DefaultMinValidators)
}