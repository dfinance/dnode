package keeper

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dfinance/dnode/x/markets/internal/types"
)

// InitGenesis inits module genesis state: creates currencies.
func (k Keeper) InitGenesis(ctx sdk.Context, data json.RawMessage) {
	state := types.GenesisState{}
	k.cdc.MustUnmarshalJSON(data, &state)

	k.SetParams(ctx, state.Params)
}

// ExportGenesis exports module genesis state using current params state.
func (k Keeper) ExportGenesis(ctx sdk.Context) json.RawMessage {
	params := k.GetParams(ctx)
	genesis := types.NewGenesisState(params)

	return k.cdc.MustMarshalJSON(genesis)
}

// InitDefaultGenesis is used for easier unit tests setup for other currencies dependant modules.
func (k Keeper) InitDefaultGenesis(ctx sdk.Context) {
	bz := k.cdc.MustMarshalJSON(types.DefaultGenesisState())
	k.InitGenesis(ctx, bz)
}
