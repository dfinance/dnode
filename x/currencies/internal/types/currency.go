package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Currency is an info object with currency params.
type Currency struct {
	// Currency denom (symbol)
	Denom string `json:"denom" example:"dfi"`
	// Number of currency decimals
	Decimals uint8 `json:"decimals" example:"0"`
	// Total amount of currency coins in Bank
	Supply sdk.Int `json:"supply" swaggertype:"string" example:"100"`
}

func (c Currency) String() string {
	return fmt.Sprintf("Currency:\n"+
		"  Denom:    %s\n"+
		"  Decimals: %d\n"+
		"  Supply:   %s",
		c.Denom,
		c.Decimals,
		c.Supply.String(),
	)
}

// NewCurrency creates a new Currency object.
func NewCurrency(denom string, supply sdk.Int, decimals uint8) Currency {
	return Currency{
		Denom:    denom,
		Decimals: decimals,
		Supply:   supply,
	}
}
