package rest

import (
	vmClient "github.com/dfinance/dnode/x/vm/client"
	"github.com/dfinance/dnode/x/vm/internal/types"
)

//nolint:deadcode,unused
type (
	VmRespCompile struct {
		Height int64             `json:"height"`
		Result vmClient.MoveFile `json:"result"`
	}

	VmData struct {
		Height int64                `json:"height"`
		Result types.QueryValueResp `json:"result" format:"HEX string"`
	}

	VmTxStatus struct {
		Height int64            `json:"height"`
		Result types.TxVMStatus `json:"result"`
	}
)
