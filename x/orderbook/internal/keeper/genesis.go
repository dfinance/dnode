package keeper

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dfinance/dnode/x/orderbook/internal/types"
)

// InitGenesis inits module genesis state: creates history items.
func (k Keeper) InitGenesis(ctx sdk.Context, data json.RawMessage) {
	k.modulePerms.AutoCheck(types.PermInit)

	state := types.GenesisState{}
	k.cdc.MustUnmarshalJSON(data, &state)

	if err := state.Validate(ctx.BlockTime(), ctx.BlockHeight()); err != nil {
		panic(err)
	}

	for _, item := range state.HistoryItems {
		k.SetHistoryItem(ctx, item)
	}
}

// ExportGenesis exports module genesis state using current params state.
func (k Keeper) ExportGenesis(ctx sdk.Context) json.RawMessage {
	k.modulePerms.AutoCheck(types.PermExport)

	state := types.GenesisState{}

	historyItems, err := k.GetHistoryItemsList(ctx)
	if err != nil {
		panic(err)
	}

	state.HistoryItems = append(state.HistoryItems, historyItems...)

	return k.cdc.MustMarshalJSON(state)
}

// InitDefaultGenesis is used for easier unit tests setup for other module dependant modules.
func (k Keeper) InitDefaultGenesis(ctx sdk.Context) {
	bz := k.cdc.MustMarshalJSON(types.DefaultGenesisState())
	k.InitGenesis(ctx, bz)
}
