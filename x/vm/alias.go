package vm

import (
	"wings-blockchain/x/vm/internal/keeper"
	"wings-blockchain/x/vm/internal/types"
	"wings-blockchain/x/vm/internal/types/vm_grpc"
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
)

var (
	NewKeeper               = keeper.NewKeeper
	RegisterVMServiceServer = vm_grpc.RegisterVMServiceServer
)
