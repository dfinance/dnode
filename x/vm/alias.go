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
	// Event types, attribute types and values
	EventTypeContractStatus = types.EventTypeContractStatus
	EventTypeMoveEvent      = types.EventTypeMoveEvent
	//
	AttributeStatus      = types.AttributeStatus
	AttributeMajorStatus = types.AttributeErrMajorStatus
	AttributeSubStatus   = types.AttributeErrSubStatus
	AttributeMessage     = types.AttributeErrMessage
	AttributeType        = types.AttributeVmEventType
	AttributeSender      = types.AttributeVmEventSender
	AttributeSource      = types.AttributeVmEventSource
	AttributeData        = types.AttributeVmEventData
	//
	AttributeValueStatusDiscard = types.AttributeValueStatusDiscard
	AttributeValueStatusKeep    = types.AttributeValueStatusKeep
	AttributeValueStatusError   = types.AttributeValueStatusError
	AttributeValueSourceScript  = types.AttributeValueSourceScript
)

type (
	Keeper           = keeper.Keeper
	MsgDeployModule  = types.MsgDeployModule
	MsgExecuteScript = types.MsgExecuteScript
	ScriptArg        = types.ScriptArg

	GenesisState    = types.GenesisState
	QueryAccessPath = types.ValueReq
	QueryValueResp  = types.ValueResp

	CurrentTimestamp = middlewares.CurrentTimestamp
	BlockHeader      = middlewares.BlockHeader

	PlannedProposal      = types.PlannedProposal
	TestProposal         = types.TestProposal
	StdlibUpdateProposal = types.StdlibUpdateProposal
)

var (
	// variable aliases
	ModuleCdc            = types.ModuleCdc
	AvailablePermissions = types.AvailablePermissions
	// function aliases
	RegisterCodec      = types.RegisterCodec
	NewKeeper          = keeper.NewKeeper
	NewMsgDeployModule = types.NewMsgDeployModule
	// error aliases
	ErrInternal           = types.ErrInternal
	ErrVMCrashed          = types.ErrVMCrashed
	ErrGovInvalidProposal = types.ErrGovInvalidProposal
)
