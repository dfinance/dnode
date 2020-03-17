package oracle

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dfinance/dnode/x/oracle/internal/types"
)

// InitGenesis sets distribution information for genesis.
func InitGenesis(ctx sdk.Context, keeper Keeper, data types.GenesisState) {

	// Set the assets and oracles from params
	keeper.SetParams(ctx, data.Params)

	// Iterate through the posted prices and set them in the store
	for _, pp := range data.PostedPrices {
		_, err := keeper.SetPrice(ctx, pp.OracleAddress, pp.AssetCode, pp.Price, pp.ReceivedAt)
		if err != nil {
			panic(err)
		}
	}

	_ = keeper.SetCurrentPrices(ctx)
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, keeper Keeper) types.GenesisState {

	// Get the params for assets and oracles
	params := keeper.GetParams(ctx)

	var postedPrices []PostedPrice
	for _, asset := range keeper.GetAssetParams(ctx) {
		pp := keeper.GetRawPrices(ctx, asset.AssetCode, ctx.BlockHeight())
		postedPrices = append(postedPrices, pp...)
	}

	return types.GenesisState{
		Params:       params,
		PostedPrices: postedPrices,
	}
}
