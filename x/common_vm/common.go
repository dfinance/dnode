package common_vm

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dfinance/dvm-proto/go/vm_grpc"
)

// Interface for other keepers to get/set data.
type VMStorage interface {
	// Access path for pricefeed.
	GetOracleAccessPath(assetCode string) *vm_grpc.VMAccessPath

	// Setters/getters.
	SetValue(ctx sdk.Context, accessPath *vm_grpc.VMAccessPath, value []byte)
	GetValue(ctx sdk.Context, accessPath *vm_grpc.VMAccessPath) []byte

	// Delete value in VM storage.
	DelValue(ctx sdk.Context, accessPath *vm_grpc.VMAccessPath)
}
