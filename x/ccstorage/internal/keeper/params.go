package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dfinance/dnode/x/ccstorage/internal/types"
)

// GetCurrenciesParams returns currencies parameters from the params storage.
func (k Keeper) GetCurrenciesParams(ctx sdk.Context) types.CurrenciesParams {
	params := types.CurrenciesParams{}
	if !k.paramStore.Has(ctx, types.ParamStoreKeyCurrencies) {
		return params
	}
	k.paramStore.Get(ctx, types.ParamStoreKeyCurrencies, &params)

	return params
}

// setCurrenciesParams sets currencies parameters (initialized from genesis) to the params storage.
func (k Keeper) setCurrenciesParams(ctx sdk.Context, params types.CurrenciesParams) {
	k.paramStore.Set(ctx, types.ParamStoreKeyCurrencies, params)
}

// updateCurrenciesParams updates currenciesParams with new (updated) currency.
func (k Keeper) updateCurrenciesParams(ctx sdk.Context, ccDenom string, ccParams types.CurrencyParams) {
	params := k.GetCurrenciesParams(ctx)
	params[ccDenom] = ccParams
	k.setCurrenciesParams(ctx, params)
}
