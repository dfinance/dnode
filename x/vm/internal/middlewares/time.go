package middlewares

import (
	"bytes"
	"encoding/hex"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dfinance/dvm-proto/go/vm_grpc"
	"github.com/dfinance/lcs"

	"github.com/dfinance/dnode/x/common_vm"
)

var (
	hexTimePath = "01bb4980dfba817aaa64cb4b3a75ee1e1d8352111dead64c5c4f05fb7b4c85bb3e"
)

type CurrentTimestamp struct {
	Seconds uint64
}

func NewTimeMiddleware() common_vm.DSDataMiddleware {
	bzPath, err := hex.DecodeString(hexTimePath)
	if err != nil {
		panic(err)
	}

	timeHeaderPath := vm_grpc.VMAccessPath{
		Address: common_vm.StdLibAddress,
		Path:    bzPath,
	}

	return func(ctx sdk.Context, path *vm_grpc.VMAccessPath) (data []byte, err error) {
		if bytes.Equal(timeHeaderPath.Address, path.Address) && bytes.Equal(timeHeaderPath.Path, path.Path) {
			return lcs.Marshal(CurrentTimestamp{Seconds: uint64(ctx.BlockHeader().Time.Unix())})
		}

		return
	}
}
