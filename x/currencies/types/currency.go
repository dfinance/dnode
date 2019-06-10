package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"fmt"
)

type Currency struct {
	Symbol   string 		`json:"symbol"`
	Supply   int64			`json:"supply"`
	Decimals int8   		`json:"decimals"`
	Creator  sdk.AccAddress `json:"creator"`
}

// New currency
func NewCurrency(symbol string, supply int64, decimals int8, creator sdk.AccAddress) Currency {
	return Currency{
		Symbol:   symbol,
		Supply:   supply,
		Decimals: decimals,
		Creator:  creator,
	}
}

func (c Currency) String() string {
	return fmt.Sprintf("Currency: \n" +
		"\tSymbol:   %s\n" +
		"\tSupply:   %d\n" +
		"\tDecimals: %d\n" +
		"\tCreator:  %s\n",
			c.Symbol, c.Supply, c.Decimals, c.Creator.String())
}