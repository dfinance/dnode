// Keeper methods related to data source.
package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"wings-blockchain/x/vm/internal/types"
)

// Start Data source (DS) server.
func (keeper *Keeper) StartDSServer(ctx sdk.Context) {
	// check if genesis initialized
	// if no - skip, it will be started later.
	store := ctx.KVStore(keeper.storeKey)
	if store.Has(types.KeyGenesisInitialized) && !keeper.dsServer.IsStarted() {
		// launch server.
		keeper.rawDSServer = StartServer(keeper.listener, keeper.dsServer)
	}
}

// Set DS (data-source) server context.
func (keeper Keeper) SetDSContext(ctx sdk.Context) {
	keeper.dsServer.SetContext(ctx)
}

// Stop DS server and close connection to VM.
func (keeper Keeper) CloseConnections() {
	if keeper.rawDSServer != nil {
		keeper.rawDSServer.Stop()
	}

	if keeper.rawClient != nil {
		keeper.rawClient.Close()
	}
}
