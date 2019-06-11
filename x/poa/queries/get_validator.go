package queries

import (
	poaTypes "wings-blockchain/x/poa/types"
	"fmt"
)

// Get validator response
type QueryGetValidatorRes struct {
	validator poaTypes.Validator `json:"validator"`
}

func (q QueryGetValidatorRes) String() string {
	return fmt.Sprintf("%s", q.validator)
}