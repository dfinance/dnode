package keeper

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dfinance/dnode/x/poa/internal/types"
)

// InitGenesis inits module genesis state: creates currencies.
func (k Keeper) InitGenesis(ctx sdk.Context, data json.RawMessage) {
	state := types.GenesisState{}
	k.cdc.MustUnmarshalJSON(data, &state)

	k.SetParams(ctx, state.Parameters)
	for _, v := range state.Validators {
		if err := k.AddValidator(ctx, v.Address, v.EthAddress); err != nil {
			panic(err)
		}
	}
}

// ExportGenesis exports module genesis state using current params state.
func (k Keeper) ExportGenesis(ctx sdk.Context) json.RawMessage {
	state := types.GenesisState{
		Parameters: k.GetParams(ctx),
		Validators: k.GetValidators(ctx),
	}

	return k.cdc.MustMarshalJSON(state)
}

// InitDefaultGenesis is used for easier unit tests setup for other module dependant modules.
func (k Keeper) InitDefaultGenesis(ctx sdk.Context) {
	bz := k.cdc.MustMarshalJSON(types.DefaultGenesisState())
	k.InitGenesis(ctx, bz)
}
