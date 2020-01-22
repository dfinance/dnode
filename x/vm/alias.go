package vm

import (
	"wings-blockchain/x/vm/internal/keeper"
	"wings-blockchain/x/vm/internal/types"
)

const (
	ModuleName        = types.ModuleName
	RouterKey         = types.RouterKey
	StoreKey          = types.StoreKey
	DefaultParamspace = types.DefaultParamspace
)

type (
	Keeper            = keeper.Keeper
	MsgDeployContract = types.MsgDeployContract
	ErrVMCrashed      = types.ErrVMCrashed
)

var (
	NewKeeper = keeper.NewKeeper
)
