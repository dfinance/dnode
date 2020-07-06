package types

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/x/params"
)

// Default parameters values.
const (
	DefaultMaxValidators uint16 = 11
	DefaultMinValidators uint16 = 3
)

// Parameter store key.
var (
	ParamStoreKeyMaxValidators = []byte("maxValidators")
	ParamStoreKeyMinValidators = []byte("minValidators")
)

// Params defines genesis params.
type Params struct {
	// Maximum number of validators allowed
	MaxValidators uint16 `json:"max_validators"`
	// Minimum number of validators allowed
	MinValidators uint16 `json:"min_validators"`
}

// Implements subspace.ParamSet interface.
func (p *Params) ParamSetPairs() params.ParamSetPairs {
	nilValidator := func(value interface{}) error { return nil }

	return params.ParamSetPairs{
		{Key: ParamStoreKeyMaxValidators, Value: &p.MaxValidators, ValidatorFn: nilValidator},
		{Key: ParamStoreKeyMinValidators, Value: &p.MinValidators, ValidatorFn: nilValidator},
	}
}

// Equal checks params equality.
func (p *Params) Equal(p2 Params) bool {
	return p.MinValidators == p2.MinValidators &&
		p.MaxValidators == p2.MaxValidators
}

// Validate validates params.
func (p Params) Validate() error {
	if p.MinValidators < DefaultMinValidators {
		return fmt.Errorf("minimum amount of validators should be not less than %d", DefaultMinValidators)
	}

	if p.MaxValidators > DefaultMaxValidators {
		return fmt.Errorf("maximum amount of validators should be not greater than %d", DefaultMaxValidators)
	}

	return nil
}

func (p Params) String() string {
	return fmt.Sprintf("Params:\n"+
		"  Max Validators: %d\n"+
		"  Min Validators: %d",
		p.MaxValidators,
		p.MinValidators,
	)
}

// NewParams creates a new module Params.
func NewParams(maxValidators, minValidators uint16) Params {
	return Params{
		MaxValidators: maxValidators,
		MinValidators: minValidators,
	}
}

// DefaultParams returns default module params.
func DefaultParams() Params {
	return NewParams(DefaultMaxValidators, DefaultMinValidators)
}

// ParamKeyTable returns Key declaration for parameters storage.
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&Params{})
}
