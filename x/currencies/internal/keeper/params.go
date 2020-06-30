package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dfinance/dnode/x/currencies/internal/types"
)

// SetCurrenciesParams sets currencies parameters (initialized from genesis) to the params storage.
func (k Keeper) SetCurrenciesParams(ctx sdk.Context, params types.CurrenciesParams) {
	k.paramStore.Set(ctx, types.ParamStoreKeyCurrencies, params)
}

// GetCurrenciesParams returns currencies parameters from the params storage.
func (k Keeper) GetCurrenciesParams(ctx sdk.Context) types.CurrenciesParams {
	params := types.CurrenciesParams{}
	k.paramStore.Get(ctx, types.ParamStoreKeyCurrencies, &params)

	return params
}
