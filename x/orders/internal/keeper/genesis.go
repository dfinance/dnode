package keeper

import (
	"encoding/json"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dfinance/dnode/x/orders/internal/types"
)

// InitGenesis inits module genesis state: creates currencies.
func (k Keeper) InitGenesis(ctx sdk.Context, data json.RawMessage) {
	k.modulePerms.AutoCheck(types.PermRead)

	state := types.GenesisState{}
	k.cdc.MustUnmarshalJSON(data, &state)
	for _, order := range state.Orders {
		k.set(ctx, order)
	}

	if err := state.Validate(ctx.BlockTime()); err != nil {
		panic(err)
	}

	if state.LastOrderId != nil {
		k.setID(ctx, *state.LastOrderId)
	}
}

// ExportGenesis exports module genesis state using current params state.
func (k Keeper) ExportGenesis(ctx sdk.Context) json.RawMessage {
	k.modulePerms.AutoCheck(types.PermRead)

	state := types.GenesisState{}

	orders, err := k.GetList(ctx)
	if err != nil {
		panic(err)
	}

	state.Orders = append(state.Orders, orders...)

	if ok := k.hasLastOrderID(ctx); ok {
		lastID := k.getLastOrderID(ctx)
		state.LastOrderId = &lastID
	}

	return k.cdc.MustMarshalJSON(state)
}

// InitDefaultGenesis is used for easier unit tests setup for other module dependant modules.
func (k Keeper) InitDefaultGenesis(ctx sdk.Context) {
	bz := k.cdc.MustMarshalJSON(types.DefaultGenesisState())
	k.InitGenesis(ctx, bz)
}
