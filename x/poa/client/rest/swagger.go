package rest

import (
	"github.com/dfinance/dnode/x/poa/types"
)

//nolint:deadcode,unused
type (
	PoaRespGetValidators struct {
		Height int64                         `json:"height"`
		Result types.ValidatorsConfirmations `json:"result"`
	}
)
