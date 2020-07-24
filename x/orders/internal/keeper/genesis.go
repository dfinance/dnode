package keeper

import (
	"encoding/json"
	"fmt"

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

	if err := state.Validate(); err != nil {
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

	iterator := k.GetIterator(ctx)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		order := types.Order{}
		if err := k.cdc.UnmarshalBinaryLengthPrefixed(iterator.Value(), &order); err != nil {
			panic(fmt.Sprintf("order unmarshal: %v", err))
		}

		state.Orders = append(state.Orders, order)
	}

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
