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
	hexBlockHeaderPath = "01ada6f79e8eddfdf986687174de1000df3c5fa45e9965ece812fed33332ec543a"
)

type BlockHeader struct {
	Height uint64
}

// NewBlockMiddleware creates DS server middleware which return current blockHeight.
func NewBlockMiddleware() common_vm.DSDataMiddleware {
	bzPath, err := hex.DecodeString(hexBlockHeaderPath)
	if err != nil {
		panic(err)
	}

	blockHeaderPath := vm_grpc.VMAccessPath{
		Address: common_vm.StdLibAddress,
		Path:    bzPath,
	}

	return func(ctx sdk.Context, path *vm_grpc.VMAccessPath) (data []byte, err error) {
		if bytes.Equal(blockHeaderPath.Address, path.Address) && bytes.Equal(blockHeaderPath.Path, path.Path) {
			return lcs.Marshal(BlockHeader{Height: uint64(ctx.BlockHeader().Height)})
		}

		return
	}
}
