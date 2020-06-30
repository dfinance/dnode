package keeper

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dfinance/dnode/x/currencies/internal/types"
)

// InitGenesis inits module genesis state: creates currencies.
func (k Keeper) InitGenesis(ctx sdk.Context, data json.RawMessage) {
	state := types.GenesisState{}
	k.cdc.MustUnmarshalJSON(data, &state)

	for denom, params := range state.CurrenciesParams {
		if err := k.CreateCurrency(ctx, denom, params); err != nil {
			panic(err)
		}
	}
}

// ExportGenesis exports module genesis state using current params state.
func (k Keeper) ExportGenesis(ctx sdk.Context) json.RawMessage {
	store := ctx.KVStore(k.storeKey)
	state := types.GenesisState{
		CurrenciesParams: make(types.CurrenciesParams, 0),
	}

	keyPrefix := types.KeyCurrencyPrefix
	keyPrefix = append(keyPrefix, types.KeyDelimiter...)
	iterator := sdk.KVStorePrefixIterator(store, keyPrefix)
	for ; iterator.Valid(); iterator.Next() {
		currency := types.Currency{}
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &currency)

		balancePath, err := k.GetCurrencyBalancePath(ctx, currency.Denom)
		if err != nil {
			panic(err)
		}

		infoPath, err := k.GetCurrencyInfoPath(ctx, currency.Denom)
		if err != nil {
			panic(err)
		}

		params := types.NewCurrencyParams(currency.Decimals, balancePath, infoPath)
		state.CurrenciesParams[currency.Denom] = params
	}

	return k.cdc.MustMarshalJSON(state)
}

// InitDefaultGenesis is used for easier unit tests setup for other currencies dependant modules.
func (k Keeper) InitDefaultGenesis(ctx sdk.Context) {
	bz := k.cdc.MustMarshalJSON(types.DefaultGenesisState())
	k.InitGenesis(ctx, bz)
}
