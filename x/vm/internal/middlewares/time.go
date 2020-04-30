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
	hexTimePath = "01843fe2742279c179713007ed25a37dc28652803c253cc3e08c75580c721060b1"
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
		Address: defaultAddr,
		Path:    bzPath,
	}

	return func(ctx sdk.Context, path *vm_grpc.VMAccessPath) (data []byte, err error) {
		if bytes.Equal(timeHeaderPath.Address, path.Address) && bytes.Equal(timeHeaderPath.Path, path.Path) {
			return lcs.Marshal(CurrentTimestamp{Seconds: uint64(ctx.BlockHeader().Time.Unix())})
		}

		return
	}
}
