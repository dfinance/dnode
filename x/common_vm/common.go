package common_vm

import (
	"bytes"
	"fmt"

	dnTypes "github.com/dfinance/dnode/helpers/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dfinance/dvm-proto/go/vm_grpc"
)

const (
	// Default address length (Move address length)
	VMAddressLength = 20
)

var (
	// Storage keys
	KeyDelimiter = []byte("@:@") // complex delimiter used, as VM accessPath.Path might include symbols like: [':', '@',..]
	VMKey        = []byte("vm")
	// Move stdlib addresses
	StdLibAddress         = make([]byte, VMAddressLength)
	StdLibAddressShortStr = "0x1"
)

// DSDataMiddleware defines prototype for DataSource server middleware.
type DSDataMiddleware func(ctx sdk.Context, path *vm_grpc.VMAccessPath) ([]byte, error)

// VMStorage interface used by other keepers to get/set VM data.
type VMStorage interface {
	// VM accessPath for an oracle asset price resource
	GetOracleAccessPath(assetCode dnTypes.AssetCode) *vm_grpc.VMAccessPath

	// Setters / getters for a VM storage values
	SetValue(ctx sdk.Context, accessPath *vm_grpc.VMAccessPath, value []byte)
	GetValue(ctx sdk.Context, accessPath *vm_grpc.VMAccessPath) []byte

	// Delete VM value from a VM storage
	DelValue(ctx sdk.Context, accessPath *vm_grpc.VMAccessPath)

	// Check value in a VM storage exists
	HasValue(ctx sdk.Context, accessPath *vm_grpc.VMAccessPath) bool
}

// GetPathKey returns storage key for VM values from VM AccessPath.
func GetPathKey(path *vm_grpc.VMAccessPath) []byte {
	// as we can't influence to path.Address/path.Path format
	// we should check if that storage key is parsable later (vm.GenesisExport)
	if bytes.Contains(path.Address, KeyDelimiter) {
		panic(fmt.Errorf("VMAccessPath.Address contains delimiter symbols"))
	}
	if bytes.Contains(path.Path, KeyDelimiter) {
		panic(fmt.Errorf("VMAccessPath.Path contains delimiter symbols"))
	}

	return bytes.Join(
		[][]byte{
			VMKey,
			path.Address,
			path.Path,
		},
		KeyDelimiter,
	)
}

// GetPathPrefixKey returns strage key prefix for VM values (used for iteration).
func GetPathPrefixKey() []byte {
	return append(VMKey, KeyDelimiter...)
}

// MustParsePathKey parses VM storage key and panics on failure.
func MustParsePathKey(key []byte) *vm_grpc.VMAccessPath {
	accessPath := vm_grpc.VMAccessPath{}

	values := bytes.Split(key, KeyDelimiter)
	if len(values) != 3 {
		panic(fmt.Errorf("key %q: invalid splitted length %d", string(key), len(values)))
	}

	if !bytes.Equal(values[0], VMKey) {
		panic(fmt.Errorf("key %q: value[0] %q: wrong prefix", string(key), string(values[0])))
	}

	accessPath.Address = values[1]
	accessPath.Path = values[2]

	return &accessPath
}

// Bech32ToLibra converts Bech32 to Libra hex.
func Bech32ToLibra(addr sdk.AccAddress) []byte {
	return addr.Bytes()
}

func init() {
	StdLibAddress[VMAddressLength-1] = 1
}
