package multisig

import (
	"github.com/dfinance/dnode/x/multisig/internal/keeper"
	"github.com/dfinance/dnode/x/multisig/internal/types"
)

type (
	Keeper            = keeper.Keeper
	GenesisState      = types.GenesisState
	Params            = types.Params
	Call              = types.Call
	Votes             = types.Votes
	MsgSubmitCall     = types.MsgSubmitCall
	MsgConfirmCall    = types.MsgConfirmCall
	MsgRevokeConfirm  = types.MsgRevokeConfirm
	CallReq           = types.CallReq
	CallByUniqueIdReq = types.CallByUniqueIdReq
	CallsResp         = types.CallsResp
	CallResp          = types.CallResp
	LastCallIdResp    = types.LastCallIdResp
)

const (
	ModuleName        = types.ModuleName
	StoreKey          = types.StoreKey
	DefaultParamspace = types.DefaultParamspace
	RouterKey         = types.RouterKey
	//
	QueryCalls        = types.QueryCalls
	QueryCall         = types.QueryCall
	QueryCallByUnique = types.QueryCallByUnique
	QueryLastId       = types.QueryLastId
	// Event types, attribute types and values
	EventTypeSubmitCall  = types.EventTypeSubmitCall
	EventTypeRemoveCall  = types.EventTypeRemoveCall
	EventTypeUpdateCall  = types.EventTypeUpdateCall
	EventTypeConfirmVote = types.EventTypeConfirmVote
	EventTypeRevokeVote  = types.EventTypeRevokeVote
	//
	AttributeMsgType   = types.AttributeMsgType
	AttributeMsgRoute  = types.AttributeMsgRoute
	AttributeCallId    = types.AttributeCallId
	AttributeUniqueId  = types.AttributeUniqueId
	AttributeSender    = types.AttributeSender
	AttributeCallState = types.AttributeCallState
	//
	AttributeValueApproved = types.AttributeValueApproved
	AttributeValueRejected = types.AttributeValueRejected
	AttributeValueFailed   = types.AttributeValueFailed
	AttributeValueExecuted = types.AttributeValueExecuted
	// Permissions
	PermPoaReader = types.PermPoaReader
	PermReader    = types.PermReader
	PermWriter    = types.PermWriter
)

var (
	// variable aliases
	ModuleCdc            = types.ModuleCdc
	AvailablePermissions = types.AvailablePermissions
	// function aliases
	RegisterCodec            = types.RegisterCodec
	NewKeeper                = keeper.NewKeeper
	NewQuerier               = keeper.NewQuerier
	DefaultGenesisState      = types.DefaultGenesisState
	NewCallStateChangedEvent = types.NewCallStateChangedEvent
	// perms requests
	RequestPoaPerms = types.RequestPoaPerms
	// errors
	ErrInternal             = types.ErrInternal
	ErrWrongCallId          = types.ErrWrongCallId
	ErrWrongCallUniqueId    = types.ErrWrongCallUniqueId
	ErrWrongMsg             = types.ErrWrongMsg
	ErrWrongMsgRoute        = types.ErrWrongMsgRoute
	ErrWrongMsgType         = types.ErrWrongMsgType
	ErrVoteAlreadyApproved  = types.ErrVoteAlreadyApproved
	ErrVoteAlreadyConfirmed = types.ErrVoteAlreadyConfirmed
	ErrVoteAlreadyRejected  = types.ErrVoteAlreadyRejected
	ErrVoteNoVotes          = types.ErrVoteNoVotes
	ErrVoteNotApproved      = types.ErrVoteNotApproved
	ErrPoaNotValidator      = types.ErrPoaNotValidator
)
