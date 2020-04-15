package vm

import (
	"github.com/dfinance/dvm-proto/go/vm_grpc"

	"github.com/dfinance/dnode/x/vm/internal/keeper"
	"github.com/dfinance/dnode/x/vm/internal/types"
)

const (
	ModuleName = types.ModuleName
	StoreKey   = types.StoreKey
)

type (
	Keeper           = keeper.Keeper
	MsgDeployModule  = types.MsgDeployModule
	MsgExecuteScript = types.MsgExecuteScript

	VMServer                     = vm_grpc.VMServiceServer
	UnimplementedVMServiceServer = vm_grpc.UnimplementedVMServiceServer

	GenesisState    = types.GenesisState
	QueryAccessPath = types.QueryAccessPath
	QueryValueResp  = types.QueryValueResp
)

var (
	NewKeeper               = keeper.NewKeeper
	RegisterVMServiceServer = vm_grpc.RegisterVMServiceServer
	MakePathKey             = types.MakePathKey

	ErrVMCrashed = types.ErrVMCrashed
)
