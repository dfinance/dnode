package middlewares

import (
	"bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dfinance/dvm-proto/go/vm_grpc"
	"github.com/dfinance/glav"
	"github.com/dfinance/lcs"

	"github.com/dfinance/dnode/x/common_vm"
)

type CurrentTimestamp struct {
	Seconds uint64
}

// NewTimeMiddleware creates DS server middleware which return current block timestamp.
func NewTimeMiddleware() common_vm.DSDataMiddleware {
	timeHeaderPath := vm_grpc.VMAccessPath{
		Address: common_vm.StdLibAddress,
		Path:    glav.TimeMetadataVector(),
	}

	return func(ctx sdk.Context, path *vm_grpc.VMAccessPath) (data []byte, err error) {
		if bytes.Equal(timeHeaderPath.Address, path.Address) && bytes.Equal(timeHeaderPath.Path, path.Path) {
			return lcs.Marshal(CurrentTimestamp{Seconds: uint64(ctx.BlockHeader().Time.Unix())})
		}

		return
	}
}
