package rest

import "github.com/dfinance/dnode/x/currencies/types"

//nolint:deadcode,unused
type (
	RespGetDestroys struct {
		Height int64          `json:"height"`
		Result types.Destroys `json:"result"`
	}

	RespGetDestroy struct {
		Height int64         `json:"height"`
		Result types.Destroy `json:"result"`
	}

	RespGetIssue struct {
		Height int64       `json:"height"`
		Result types.Issue `json:"result"`
	}

	RespGetCurrency struct {
		Height int64          `json:"height"`
		Result types.Currency `json:"result"`
	}
)
