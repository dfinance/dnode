package currencies

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Currency struct {
	Symbol   string 		`json:"symbol"`
	Supply   int64			`json:"supply"`
	Decimals int8   		`json:"decimals"`
	Creator  sdk.AccAddress `json:"creator"`
}

func NewCurrency(symbol string, supply int64, decimals int8, creator sdk.AccAddress) Currency {
	return Currency{
		Symbol:   symbol,
		Supply:   supply,
		Decimals: decimals,
		Creator:  creator,
	}
}