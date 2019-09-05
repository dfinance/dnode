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
func NewCurrency(currencyId sdk.Int, symbol string, supply sdk.Int, decimals int8) Currency {
	return Currency{
		CurrencyId: currencyId,
		Symbol:     symbol,
		Supply:     supply,
		Decimals:   decimals,
	}
}

func (c Currency) String() string {
	return fmt.Sprintf("Currency: \n" +
		"\tCurrency id: %s\n" +
		"\tSymbol:      %s\n" +
		"\tSupply:      %s\n" +
		"\tDecimals:    %d\n",
			c.CurrencyId.String(), c.Symbol, c.Supply.String(), c.Decimals)
}
