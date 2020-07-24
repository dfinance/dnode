package rest

import (
	"github.com/cosmos/cosmos-sdk/x/auth"

	"github.com/dfinance/dnode/x/ccstorage"
	"github.com/dfinance/dnode/x/currencies/internal/types"
)

//nolint:deadcode,unused
type (
	CCRespGetWithdraws struct {
		Height int64           `json:"height"`
		Result types.Withdraws `json:"result"`
	}

	CCRespGetWithdraw struct {
		Height int64          `json:"height"`
		Result types.Withdraw `json:"result"`
	}

	CCRespGetIssue struct {
		Height int64       `json:"height"`
		Result types.Issue `json:"result"`
	}

	CCRespGetCurrency struct {
		Height int64              `json:"height"`
		Result ccstorage.Currency `json:"result"`
	}

	CCRespGetCurrencies struct {
		Height int64                `json:"height"`
		Result ccstorage.Currencies `json:"result"`
	}

	CCRespStdTx struct {
		Height int64      `json:"height"`
		Result auth.StdTx `json:"result"`
	}
)
