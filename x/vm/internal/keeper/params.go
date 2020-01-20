package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	"time"
	types "wings-blockchain/x/vm/internal/types"
)

func NewKeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&types.Params{})
}

func (keeper Keeper) GetVMAddress(ctx sdk.Context) (vmAddress string) {
	keeper.paramStore.Get(ctx, types.KeyVMAddress, &vmAddress)
	return
}

func (keeper Keeper) GetVMTimeout(ctx sdk.Context) (vmTimeout time.Duration) {
	keeper.paramStore.Get(ctx, types.KeyVMTimeout, &vmTimeout)
	return
}

func (keeper Keeper) GetParams(ctx sdk.Context) types.Params {
	vmAddress := keeper.GetVMAddress(ctx)
	vmTimeout := keeper.GetVMTimeout(ctx)

	return types.NewParams(vmAddress, vmTimeout)
}

func (keeper Keeper) SetParams(ctx sdk.Context, params types.Params) {
	keeper.paramStore.SetParamSet(ctx, &params)
}
