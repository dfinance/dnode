package queries

import (
	poaTypes "github.com/WingsDao/wings-blockchain/x/poa/types"
)

// Get validator response
type QueryGetValidatorRes struct {
	Validator poaTypes.Validator `json:"validator"`
}

func (q QueryGetValidatorRes) String() string {
	return q.Validator.String()
}
