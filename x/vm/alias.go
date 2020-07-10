package vm

import (
	"github.com/dfinance/dnode/x/vm/internal/keeper"
	"github.com/dfinance/dnode/x/vm/internal/middlewares"
	"github.com/dfinance/dnode/x/vm/internal/types"
)

const (
	ModuleName   = types.ModuleName
	StoreKey     = types.StoreKey
	RouterKey    = types.RouterKey
	GovRouterKey = types.GovRouterKey
)

type (
	Keeper           = keeper.Keeper
	MsgDeployModule  = types.MsgDeployModule
	MsgExecuteScript = types.MsgExecuteScript
	ScriptArg        = types.ScriptArg

	GenesisState    = types.GenesisState
	QueryAccessPath = types.QueryAccessPath
	QueryValueResp  = types.QueryValueResp

	CurrentTimestamp = middlewares.CurrentTimestamp
	BlockHeader      = middlewares.BlockHeader

	PlannedProposal      = types.PlannedProposal
	TestProposal         = types.TestProposal
	StdlibUpdateProposal = types.StdlibUpdateProposal
)

var (
	// variable aliases
	ModuleCdc = types.ModuleCdc
	// function aliases
	RegisterCodec      = types.RegisterCodec
	NewKeeper          = keeper.NewKeeper
	NewMsgDeployModule = types.NewMsgDeployModule
	// error aliases
	ErrInternal           = types.ErrInternal
	ErrVMCrashed          = types.ErrVMCrashed
	ErrGovInvalidProposal = types.ErrGovInvalidProposal
)
