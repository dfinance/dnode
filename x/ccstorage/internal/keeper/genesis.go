package keeper

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dfinance/dnode/x/ccstorage/internal/types"
)

// InitGenesis inits module genesis state: creates currencies.
func (k Keeper) InitGenesis(ctx sdk.Context, data json.RawMessage) {
	k.modulePerms.AutoCheck(types.PermInit)

	state := types.GenesisState{}
	k.cdc.MustUnmarshalJSON(data, &state)

	for _, params := range state.CurrenciesParams {
		if err := k.CreateCurrency(ctx, params); err != nil {
			panic(err)
		}
	}
}

// ExportGenesis exports module genesis state using current params state.
func (k Keeper) ExportGenesis(ctx sdk.Context) json.RawMessage {
	k.modulePerms.AutoCheck(types.PermRead)

	state := types.GenesisState{
		CurrenciesParams: types.CurrenciesParams{},
	}

	for _, currency := range k.GetCurrencies(ctx) {
		state.CurrenciesParams = append(state.CurrenciesParams, types.CurrencyParams{
			Denom:    currency.Denom,
			Decimals: currency.Decimals,
		})
	}

	return k.cdc.MustMarshalJSON(state)
}

// InitDefaultGenesis is used for easier unit tests setup for other currencies dependant modules.
func (k Keeper) InitDefaultGenesis(ctx sdk.Context) {
	bz := k.cdc.MustMarshalJSON(types.DefaultGenesisState())
	k.InitGenesis(ctx, bz)
}
