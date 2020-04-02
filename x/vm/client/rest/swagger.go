package rest

import vmClient "github.com/dfinance/dnode/x/vm/client"

//nolint:deadcode,unused
type (
	VmRespCompile struct {
		Height int64           `json:"height"`
		Result vmClient.MVFile `json:"result"`
	}
)
