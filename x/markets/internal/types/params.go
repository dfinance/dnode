package types

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/x/params/subspace"
)

// Storage keys.
var (
	KeyMarkets = []byte("markets")
)

// Keeper params type.
type Params struct {
	Markets Markets
}

// Implements subspace.ParamSet.
func (p *Params) ParamSetPairs() subspace.ParamSetPairs {
	nilPairValidatorFunc := func(value interface{}) error {
		return nil
	}

	return subspace.ParamSetPairs{
		subspace.NewParamSetPair(KeyMarkets, &p.Markets, nilPairValidatorFunc),
	}
}

// Validate validates keeper params.
func (p Params) Validate() error {
	for i, m := range p.Markets {
		if err := m.Valid(); err != nil {
			return fmt.Errorf("market [%d] %s: %v", i, m.String(), err)
		}

		if m.ID.UInt64() != uint64(i) {
			return fmt.Errorf("market [%d] %s: invalid ID (params order mismatch)", i, m.String())
		}
	}

	return nil
}

// NewParams creates a new keeper params object.
func NewParams(markets []Market, nominees []string) Params {
	return Params{
		Markets: markets,
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
