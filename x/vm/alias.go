package vm

import (
	"github.com/dfinance/dvm-proto/go/vm_grpc"

	"github.com/dfinance/dnode/x/vm/internal/keeper"
	"github.com/dfinance/dnode/x/vm/internal/types"
)

const (
	ModuleName   = types.ModuleName
	StoreKey     = types.StoreKey
	GovRouterKey = types.GovRouterKey
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

	PlannedProposal      = types.PlannedProposal
	ModuleUpdateProposal = types.ModuleUpdateProposal
	ModuleUpdateData     = types.ModuleUpdateData
	TestProposal         = types.TestProposal
	TestData             = types.TestData
)

var (
	NewKeeper               = keeper.NewKeeper
	RegisterVMServiceServer = vm_grpc.RegisterVMServiceServer

	ErrVMCrashed = types.ErrVMCrashed

	NewPlan                 = types.NewPlan
	NewPlannedProposal      = types.NewPlannedProposal
	NewModuleUpdateProposal = types.NewModuleUpdateProposal
	NewTestProposal         = types.NewTestProposal
)
