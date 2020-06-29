package rest

import (
	types2 "github.com/dfinance/dnode/x/currencies/internal/types"
)

//nolint:deadcode,unused
type (
	CCRespGetDestroys struct {
		Height int64           `json:"height"`
		Result types2.Destroys `json:"result"`
	}

	CCRespGetDestroy struct {
		Height int64          `json:"height"`
		Result types2.Destroy `json:"result"`
	}

	CCRespGetIssue struct {
		Height int64        `json:"height"`
		Result types2.Issue `json:"result"`
	}

	CCRespGetCurrency struct {
		Height int64           `json:"height"`
		Result types2.Currency `json:"result"`
	}
)
