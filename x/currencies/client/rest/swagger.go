package rest

import (
	"github.com/dfinance/dnode/x/currencies/internal/types"
)

//nolint:deadcode,unused
type (
	CCRespGetWithdraws struct {
		Height int64            `json:"height"`
		Result types.Withdraws `json:"result"`
	}

	CCRespGetWithdraw struct {
		Height int64           `json:"height"`
		Result types.Withdraw `json:"result"`
	}

	CCRespGetIssue struct {
		Height int64        `json:"height"`
		Result types.Issue `json:"result"`
	}

	CCRespGetCurrency struct {
		Height int64           `json:"height"`
		Result types.Currency `json:"result"`
	}
)
