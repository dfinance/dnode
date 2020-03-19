package oracle

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dfinance/dnode/x/oracle/internal/types"
)

// InitGenesis sets distribution information for genesis.
func InitGenesis(ctx sdk.Context, keeper Keeper, data types.GenesisState) {
	// Set the assets and oracles from params
	keeper.SetParams(ctx, data.Params)

	// Just adding assets.
	for _, asset := range data.Assets {
		keeper.AddAsset(ctx, data.Params.Nominees[0], asset.AssetCode, asset)
	}
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, keeper Keeper) types.GenesisState {
	// Get the params for assets and oracles
	params := keeper.GetParams(ctx)
	assets := keeper.GetAssetParams(ctx)

	return types.GenesisState{
		Params: params,
		Assets: assets,
	}
}
