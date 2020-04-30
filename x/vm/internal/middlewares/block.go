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
	hexBlockHeaderPath = "01bfb408926a938924ad7d972494ec9e2a47dc71a0b1785c817260739fcaac201d"
)

type BlockHeader struct {
	Height uint64
}

func NewBlockMiddleware() common_vm.DSDataMiddleware {
	bzPath, err := hex.DecodeString(hexBlockHeaderPath)
	if err != nil {
		panic(err)
	}

	blockHeaderPath := vm_grpc.VMAccessPath{
		Address: defaultAddr,
		Path:    bzPath,
	}

	return func(ctx sdk.Context, path *vm_grpc.VMAccessPath) (data []byte, err error) {
		if bytes.Equal(blockHeaderPath.Address, path.Address) && bytes.Equal(blockHeaderPath.Path, path.Path) {
			return lcs.Marshal(BlockHeader{Height: uint64(ctx.BlockHeader().Height)})
		}

		return
	}
}
