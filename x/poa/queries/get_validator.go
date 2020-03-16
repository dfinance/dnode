package queries

import (
	poaTypes "github.com/dfinance/dnode/x/poa/types"
)

// Get validator response
type QueryGetValidatorRes struct {
	Validator poaTypes.Validator `json:"validator"`
}

func (q QueryGetValidatorRes) String() string {
	return q.Validator.String()
}
