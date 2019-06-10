package queries

import (
	"wings-blockchain/x/currencies/types"
	"fmt"
)

// Get currency query response
type QueryCurrencyRes struct {
	Currency types.Currency `json:"currency"`
}

func (q QueryCurrencyRes) String() string {
	return fmt.Sprintf("%s", q.Currency.String())
}
