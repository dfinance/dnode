package queries

import (
	poaTypes "wings-blockchain/x/poa/types"
	"fmt"
)

// Get validator response
type QueryGetValidatorRes struct {
	validator poaTypes.Validator
}

func (q QueryGetValidatorRes) String() string {
	return fmt.Sprintf("%s", q.validator)
}