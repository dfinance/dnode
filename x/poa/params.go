package poa

import (
	"github.com/cosmos/cosmos-sdk/x/params"
	"wings-blockchain/x/poa/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewKeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&types.Params{})
}

func (keeper Keeper) GetMaxValidators(ctx sdk.Context) (res uint16) {
	keeper.paramStore.Get(ctx, types.KeyMaxValidators, &res)
	return
}

func (keeper Keeper) GetMinValidators(ctx sdk.Context) (res uint16) {
	keeper.paramStore.Get(ctx, types.KeyMinValidators, &res)
	return
}