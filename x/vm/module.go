package vm

import (
	"encoding/json"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"
	"wings-blockchain/x/vm/client/cli"
	types "wings-blockchain/x/vm/internal/types"
)

var (
	_ AppModule      = AppModule{}
	_ AppModuleBasic = AppModuleBasic{}
)

type AppModuleBasic struct{}

// Module name.
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// Registering codecs.
func (module AppModuleBasic) RegisterCodec(cdc *codec.Codec) {
	types.RegisterCodec(cdc)
}

// Validate exists genesis.
func (AppModuleBasic) ValidateGenesis(json.RawMessage) error {
	return nil
}

// Generate default genesis.
func (module AppModuleBasic) DefaultGenesis() json.RawMessage {
	return types.ModuleCdc.MustMarshalJSON(types.GenesisState{
		Parameters: types.DefaultParams(),
	})
}

// Register REST routes.
func (AppModuleBasic) RegisterRESTRoutes(ctx context.CLIContext, r *mux.Router) {}

// Get transaction commands for CLI.
func (AppModuleBasic) GetTxCmd(cdc *codec.Codec) *cobra.Command {
	return cli.GetTxCmd(cdc)
}

// Get query commands for CLI.
func (AppModuleBasic) GetQueryCmd(cdc *codec.Codec) *cobra.Command {
	return nil
}

// VM module.
type AppModule struct {
	AppModuleBasic
	vmKeeper Keeper
}

// Create new VM module.
func NewAppModule(vmKeeper Keeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		vmKeeper:       vmKeeper,
	}
}

// Get name of module.
func (AppModule) Name() string {
	return types.ModuleName
}

// Register module invariants.
func (AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}

// Base route of module (for handler).
func (AppModule) Route() string { return types.RouterKey }

// Create new handler.
func (app AppModule) NewHandler() sdk.Handler { return NewHandler(app.vmKeeper) }

// Get route for querier.
func (AppModule) QuerierRoute() string { return types.RouterKey }

// Get new querier for VM module.
func (app AppModule) NewQuerierHandler() sdk.Querier {
	return nil
}

// Process begin block (abci).
func (AppModule) BeginBlock(_ sdk.Context, _ abci.RequestBeginBlock) {}

// Process end block (abci).
func (AppModule) EndBlock(_ sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}

// Initialize genesis.
func (app AppModule) InitGenesis(ctx sdk.Context, data json.RawMessage) []abci.ValidatorUpdate {
	var genesisState types.GenesisState

	types.ModuleCdc.MustUnmarshalJSON(data, &genesisState)
	app.vmKeeper.SetParams(ctx, genesisState.Parameters)

	return []abci.ValidatorUpdate{}
}

// Export genesis.
func (app AppModule) ExportGenesis(ctx sdk.Context) json.RawMessage {
	genesisState := types.GenesisState{
		Parameters: app.vmKeeper.GetParams(ctx),
	}

	return types.ModuleCdc.MustMarshalJSON(genesisState)
}
