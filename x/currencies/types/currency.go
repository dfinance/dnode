// Currency type implementation.
package types

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Currency struct {
	CurrencyId sdk.Int `json:"currencyId"`
	Symbol     string  `json:"symbol"`
	Supply     sdk.Int `json:"supply"`
	Decimals   int8    `json:"decimals"`
}

// New currency
func NewCurrency(symbol string, supply sdk.Int, decimals int8) Currency {
	return Currency{
		Symbol:   symbol,
		Supply:   supply,
		Decimals: decimals,
	}
}

func (c Currency) String() string {
	return fmt.Sprintf("Currency: \n"+
		"\tSymbol:      %s\n"+
		"\tSupply:      %s\n"+
		"\tDecimals:    %d\n",
		c.Symbol, c.Supply.String(), c.Decimals)
}
