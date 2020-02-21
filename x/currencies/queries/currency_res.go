package queries

import (
	"github.com/WingsDao/wings-blockchain/x/currencies/types"
)

// Get currency query response
type QueryCurrencyRes struct {
	Currency types.Currency `json:"currency"`
}

func (q QueryCurrencyRes) String() string {
	return q.Currency.String()
}
