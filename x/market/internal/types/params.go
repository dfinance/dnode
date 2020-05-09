package types

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/x/params/subspace"
)

// Storage keys.
var (
	KeyMarkets  = []byte("marketMarkets")
	KeyNominees = []byte("marketNominees")
)

// Keeper params type.
type Params struct {
	Markets  Markets
	Nominees []string
}

// Implements subspace.ParamSet.
func (p *Params) ParamSetPairs() subspace.ParamSetPairs {
	nilPairValidatorFunc := func(value interface{}) error {
		return nil
	}

	return subspace.ParamSetPairs{
		subspace.NewParamSetPair(KeyMarkets, &p.Markets, nilPairValidatorFunc),
		subspace.NewParamSetPair(KeyNominees, &p.Nominees, nilPairValidatorFunc),
	}
}

// Validate validates keeper params.
func (p Params) Validate() error {
	for i, m := range p.Markets {
		if err := m.Valid(); err != nil {
			return fmt.Errorf("market [%d] %s: %v", i, m.String(), err)
		}
	}

	for i, n := range p.Nominees {
		if n == "" {
			return fmt.Errorf("nominee [%d]: empty", i)
		}
	}

	return nil
}

// NewParams creates a new keeper params object.
func NewParams(markets []Market, nominees []string) Params {
	return Params{
		Markets:  markets,
		Nominees: nominees,
	}
}

// DefaultParams returns default keeper params.
func DefaultParams() Params {
	return NewParams(Markets{}, []string{})
}

// ParamKeyTable creates keeper params KeyTable.
func ParamKeyTable() subspace.KeyTable {
	return subspace.NewKeyTable().RegisterParamSet(&Params{})
}
