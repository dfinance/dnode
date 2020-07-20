package poa

import (
	"github.com/dfinance/dnode/x/poa/internal/keeper"
	"github.com/dfinance/dnode/x/poa/internal/types"
)

type (
	Keeper                      = keeper.Keeper
	GenesisState                = types.GenesisState
	Params                      = types.Params
	Validator                   = types.Validator
	Validators                  = types.Validators
	MsgAddValidator             = types.MsgAddValidator
	MsgReplaceValidator         = types.MsgReplaceValidator
	MsgRemoveValidator          = types.MsgRemoveValidator
	ValidatorReq                = types.ValidatorReq
	ValidatorsConfirmationsResp = types.ValidatorsConfirmationsResp
)

const (
	ModuleName        = types.ModuleName
	StoreKey          = types.ModuleName
	DefaultParamspace = types.DefaultParamspace
	RouterKey         = types.RouterKey
	//
	QueryValidators = types.QueryValidators
	QueryValidator  = types.QueryValidator
	QueryMinMax     = types.QueryMinMax
	//
	DefaultMaxValidators = types.DefaultMaxValidators
	DefaultMinValidators = types.DefaultMinValidators
	// Event types, attribute types and values
	EventTypeAdd    = types.EventTypeAdd
	EventTypeRemove = types.EventTypeRemove
	//
	AttributeSdkAddress = types.AttributeSdkAddress
	AttributeEthAddress = types.AttributeEthAddress
)

var (
	// variable aliases
	ModuleCdc            = types.ModuleCdc
	AvailablePermissions = types.AvailablePermissions
	// function aliases
	RegisterCodec          = types.RegisterCodec
	NewKeeper              = keeper.NewKeeper
	NewQuerier             = keeper.NewQuerier
	DefaultParams          = types.DefaultParams
	DefaultGenesisState    = types.DefaultGenesisState
	NewMsgAddValidator     = types.NewMsgAddValidator
	NewMsgReplaceValidator = types.NewMsgReplaceValidator
	NewMsgRemoveValidator  = types.NewMsgRemoveValidator
	// errors
	ErrInternal             = types.ErrInternal
	ErrWrongEthereumAddress = types.ErrWrongEthereumAddress
	ErrValidatorExists      = types.ErrValidatorExists
	ErrValidatorNotExists   = types.ErrValidatorNotExists
	ErrMaxValidatorsReached = types.ErrMaxValidatorsReached
	ErrMinValidatorsReached = types.ErrMinValidatorsReached
)
