package types

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/x/params/subspace"
)

var (
	KeyMarkets  = []byte("marketMarkets")
	KeyNominees = []byte("marketNominees")
)

type Params struct {
	Markets  Markets
	Nominees []string
}

func (p *Params) ParamSetPairs() subspace.ParamSetPairs {
	nilPairValidatorFunc := func(value interface{}) error {
		return nil
	}

	return subspace.ParamSetPairs{
		subspace.NewParamSetPair(KeyMarkets, &p.Markets, nilPairValidatorFunc),
		subspace.NewParamSetPair(KeyNominees, &p.Nominees, nilPairValidatorFunc),
	}
}

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

func NewParams(markets []Market, nominees []string) Params {
	return Params{
		Markets:  markets,
		Nominees: nominees,
	}
}

func DefaultParams() Params {
	return NewParams(Markets{}, []string{})
}

func ParamKeyTable() subspace.KeyTable {
	return subspace.NewKeyTable().RegisterParamSet(&Params{})
}
