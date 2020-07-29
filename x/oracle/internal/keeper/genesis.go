package keeper

import (
	"encoding/json"
	"fmt"

	"github.com/dfinance/dnode/x/oracle/internal/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis inits module genesis state: creates currencies.
func (k Keeper) InitGenesis(ctx sdk.Context, data json.RawMessage) {
	k.modulePerms.AutoCheck(types.PermInit)

	state := types.GenesisState{}
	k.cdc.MustUnmarshalJSON(data, &state)

	if err := state.Validate(ctx.BlockTime()); err != nil {
		panic(err)
	}

	k.SetParams(ctx, state.Params)

	for _, cPrice := range state.CurrentPrices {
		if _, ok := k.GetAsset(ctx, cPrice.AssetCode); !ok {
			panic(fmt.Errorf("asset_code %s does not exist", cPrice.AssetCode))
		}
		k.AddCurrentPrice(ctx, cPrice)
	}
}

// ExportGenesis exports module genesis state using current params state.
func (k Keeper) ExportGenesis(ctx sdk.Context) json.RawMessage {
	k.modulePerms.AutoCheck(types.PermRead)

	currentPrices, err := k.GetCurrentPricesList(ctx)
	if err != nil {
		panic(err)
	}

	state := types.GenesisState{
		Params:        k.GetParams(ctx),
		CurrentPrices: currentPrices,
	}

	return k.cdc.MustMarshalJSON(state)
}

// InitDefaultGenesis is used for easier unit tests setup for other module dependant modules.
func (k Keeper) InitDefaultGenesis(ctx sdk.Context) {
	bz := k.cdc.MustMarshalJSON(types.DefaultGenesisState())
	k.InitGenesis(ctx, bz)
}
