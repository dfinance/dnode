package common_vm

import (
	"bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dfinance/dvm-proto/go/vm_grpc"
)

var (
	KeyDelimiter = []byte(":")
	VMKey        = []byte("vm")
)

// Interface for other keepers to get/set data.
type VMStorage interface {
	GetOracleAccessPath(assetCode string) *vm_grpc.VMAccessPath
	SetValue(ctx sdk.Context, accessPath *vm_grpc.VMAccessPath, value []byte)
	GetValue(ctx sdk.Context, accessPath *vm_grpc.VMAccessPath) []byte
	DelValue(ctx sdk.Context, accessPath *vm_grpc.VMAccessPath)
}

// Convert bech32 to libra hex.
func Bech32ToLibra(addr sdk.AccAddress) []byte {
	return append(addr, make([]byte, 4)...)
}

// Make path for storage from VMAccessPath.
func MakePathKey(path vm_grpc.VMAccessPath) []byte {
	return bytes.Join(
		[][]byte{
			VMKey,
			path.Address,
			path.Path,
		},
		KeyDelimiter,
	)
}
