package keeper

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dfinance/dnode/x/markets/internal/types"
)

// InitGenesis inits module genesis state: creates currencies.
func (k Keeper) InitGenesis(ctx sdk.Context, data json.RawMessage) {
	k.modulePerms.AutoCheck(types.PermInit)

	state := types.GenesisState{}
	k.cdc.MustUnmarshalJSON(data, &state)

	// lastMarketID
	if state.LastMarketID != nil {
		k.setLastID(ctx, *state.LastMarketID)
	}

	// markets
	{
		for i, market := range state.Markets {
			if !k.ccsStorage.HasCurrency(ctx, market.BaseAssetDenom) {
				panic(fmt.Errorf("market[%d]: baseAsset currency not found", i))
			}
			if !k.ccsStorage.HasCurrency(ctx, market.QuoteAssetDenom) {
				panic(fmt.Errorf("market[%d]: quoteAsset currency not found", i))
			}

			k.set(ctx, market)
		}
	}
}

// ExportGenesis exports module genesis state using current params state.
func (k Keeper) ExportGenesis(ctx sdk.Context) json.RawMessage {
	state := types.GenesisState{
		Markets:      k.GetList(ctx),
		LastMarketID: k.getLastMarketID(ctx),
	}

	return k.cdc.MustMarshalJSON(state)
}

// InitDefaultGenesis is used for easier unit tests setup for other currencies dependant modules.
func (k Keeper) InitDefaultGenesis(ctx sdk.Context) {
	bz := k.cdc.MustMarshalJSON(types.DefaultGenesisState())
	k.InitGenesis(ctx, bz)
}
