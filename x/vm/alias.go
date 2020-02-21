package vm

import (
	"github.com/WingsDao/wings-blockchain/x/vm/internal/keeper"
	"github.com/WingsDao/wings-blockchain/x/vm/internal/types"
	"github.com/WingsDao/wings-blockchain/x/vm/internal/types/vm_grpc"
)

const (
	ModuleName        = types.ModuleName
	RouterKey         = types.RouterKey
	StoreKey          = types.StoreKey
	DefaultParamspace = types.DefaultParamspace
)

type (
	Keeper           = keeper.Keeper
	MsgDeployModule  = types.MsgDeployModule
	MsgExecuteScript = types.MsgExecuteScript
	ErrVMCrashed     = types.ErrVMCrashed

	VMServer                     = vm_grpc.VMServiceServer
	UnimplementedVMServiceServer = vm_grpc.UnimplementedVMServiceServer

	QueryAccessPath = types.QueryAccessPath
	QueryValueResp  = types.QueryValueResp
)

var (
	NewKeeper               = keeper.NewKeeper
	RegisterVMServiceServer = vm_grpc.RegisterVMServiceServer
)
