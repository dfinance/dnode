// Parameters key table implementation for multisig parameters store.
package multisig

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"

	"github.com/dfinance/dnode/x/multisig/types"
)

// New Paramstore for multisig module.
func NewKeyTable() params.KeyTable {
	return params.NewKeyTable().RegisterParamSet(&types.Params{})
}

// Get IntervalToExecute calls parameter.
func (keeper Keeper) GetIntervalToExecute(ctx sdk.Context) (res int64) {
	keeper.paramStore.Get(ctx, types.KeyIntervalToExecute, &res)
	return
}

// Get params.
func (keeper Keeper) GetParams(ctx sdk.Context) types.Params {
	intervalToExecute := keeper.GetIntervalToExecute(ctx)

	return types.NewParams(intervalToExecute)
}

// Set the params.
func (keeper Keeper) SetParams(ctx sdk.Context, params types.Params) {
	keeper.paramStore.SetParamSet(ctx, &params)
}
