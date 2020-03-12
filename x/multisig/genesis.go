package multisig

import (
	"github.com/WingsDao/wings-blockchain/x/multisig/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Initialize genesis for this module.
func (keeper Keeper) InitGenesis(ctx sdk.Context, genesisState types.GenesisState) {
	keeper.SetParams(ctx, genesisState.Parameters)
}

// Export genesis data for this module.
func (keeper Keeper) ExportGenesis(ctx sdk.Context) types.GenesisState {
	return types.GenesisState{
		Parameters: keeper.GetParams(ctx),
	}
}
