package queries

import (
	types "wings-blockchain/x/currencies/types"
	"fmt"
)

// Get denoms query response
type QueryDenomsRes struct {
	Denoms types.Denoms	`json:"denoms"`
}

func (q QueryDenomsRes) String() string {
	return fmt.Sprintf("%v", q.Denoms)
}