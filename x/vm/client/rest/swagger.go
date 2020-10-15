package rest

import (
	"github.com/cosmos/cosmos-sdk/x/auth"

	"github.com/dfinance/dnode/x/vm/client/vm_client"
	"github.com/dfinance/dnode/x/vm/internal/types"
)

//nolint:deadcode,unused
type (
	VmRespCompile struct {
		Height int64                   `json:"height"`
		Result vm_client.CompiledItems `json:"result"`
	}

	VmData struct {
		Height int64           `json:"height"`
		Result types.ValueResp `json:"result" format:"HEX string"`
	}

	VmTxStatus struct {
		Height int64            `json:"height"`
		Result types.TxVMStatus `json:"result"`
	}

	VmRespStdTx struct {
		Height int64      `json:"height"`
		Result auth.StdTx `json:"result"`
	}

	VmRespLcsView struct {
		Height int64       `json:"height"`
		Result LcsViewResp `json:"result"`
	}
)
