package rest

import "github.com/dfinance/dnode/x/multisig/types"

//nolint:deadcode,unused
type (
	MSRespGetCall struct {
		Height int64          `json:"height"`
		Result types.CallResp `json:"result"`
	}

	MSRespGetCalls struct {
		Height int64           `json:"height"`
		Result types.CallsResp `json:"result"`
	}
)
