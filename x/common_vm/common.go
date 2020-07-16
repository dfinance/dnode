package common_vm

import (
	"bytes"
	dnTypes "github.com/dfinance/dnode/helpers/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dfinance/dvm-proto/go/vm_grpc"
)

const (
	// Default address length.
	VMAddressLength = 20
)

var (
	KeyDelimiter          = []byte(":")
	VMKey                 = []byte("vm")
	StdLibAddress         = make([]byte, VMAddressLength)
	StdLibAddressShortStr = "0x1"
)

// Data server middleware type.
type DSDataMiddleware func(ctx sdk.Context, path *vm_grpc.VMAccessPath) ([]byte, error)

// Interface for other keepers to get/set data.
type VMStorage interface {
	// Access path for oracle.
	GetOracleAccessPath(assetCode dnTypes.AssetCode) *vm_grpc.VMAccessPath

	// Setters/getters.
	SetValue(ctx sdk.Context, accessPath *vm_grpc.VMAccessPath, value []byte)
	GetValue(ctx sdk.Context, accessPath *vm_grpc.VMAccessPath) []byte

	// Delete value in VM storage.
	DelValue(ctx sdk.Context, accessPath *vm_grpc.VMAccessPath)

	// Has value in VM storage.
	HasValue(ctx sdk.Context, accessPath *vm_grpc.VMAccessPath) bool
}

// Make path for storage from VMAccessPath.
func MakePathKey(path *vm_grpc.VMAccessPath) []byte {
	return bytes.Join(
		[][]byte{
			VMKey,
			path.Address,
			path.Path,
		},
		KeyDelimiter,
	)
}

// Convert bech32 to libra hex.
func Bech32ToLibra(addr sdk.AccAddress) []byte {
	return addr.Bytes()
}

func init() {
	StdLibAddress[VMAddressLength-1] = 1
}
