package rest

import "github.com/dfinance/dnode/x/currencies/types"

//nolint:deadcode,unused
type (
	CCRespGetDestroys struct {
		Height int64          `json:"height"`
		Result types.Destroys `json:"result"`
	}

	CCRespGetDestroy struct {
		Height int64         `json:"height"`
		Result types.Destroy `json:"result"`
	}

	CCRespGetIssue struct {
		Height int64       `json:"height"`
		Result types.Issue `json:"result"`
	}

	CCRespGetCurrency struct {
		Height int64          `json:"height"`
		Result types.Currency `json:"result"`
	}
)
