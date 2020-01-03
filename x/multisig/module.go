package multisig

import (
	"encoding/json"
	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	"github.com/tendermint/go-amino"
	abci "github.com/tendermint/tendermint/abci/types"
	"wings-blockchain/x/multisig/client"
	"wings-blockchain/x/multisig/client/rest"
	"wings-blockchain/x/multisig/types"
	"wings-blockchain/x/poa"
)

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

type AppModuleBasic struct{}

// Module name.
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// Registering codecs.
func (module AppModuleBasic) RegisterCodec(cdc *amino.Codec) {
	RegisterCodec(cdc)
}

// Validate exists genesis.
func (AppModuleBasic) ValidateGenesis(bz json.RawMessage) error {
	return nil
}

// Generate default genesis.
func (module AppModuleBasic) DefaultGenesis() json.RawMessage {
	return json.RawMessage{}
}

// Register REST routes.
func (AppModuleBasic) RegisterRESTRoutes(ctx context.CLIContext, r *mux.Router) {
	rest.RegisterRoutes(ctx, r)
}

// Get transaction commands for CLI.
func (AppModuleBasic) GetTxCmd(cdc *amino.Codec) *cobra.Command {
	return client.GetTxCmd(cdc)
}

// Get query commands for CLI.
func (AppModuleBasic) GetQueryCmd(cdc *amino.Codec) *cobra.Command {
	return client.GetQueryCmd(cdc)
}

type AppModule struct {
	AppModuleBasic
	msKeeper  Keeper
	poaKeeper poa.Keeper
}

// Create new PoA module.
func NewAppModule(msKeeper Keeper, poaKeeper poa.Keeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		msKeeper:       msKeeper,
		poaKeeper:      poaKeeper,
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
func (app AppModule) NewHandler() sdk.Handler { return NewHandler(app.msKeeper, app.poaKeeper) }

// Get route for querier.
func (AppModule) QuerierRoute() string { return types.RouterKey }

// Get new querier for PoA module.
func (app AppModule) NewQuerierHandler() sdk.Querier {
	return NewQuerier(app.msKeeper)
}

// Process begin block (abci).
func (AppModule) BeginBlock(_ sdk.Context, _ abci.RequestBeginBlock) {
}

// Process end block (abci).
func (app AppModule) EndBlock(ctx sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	EndBlocker(ctx, app.msKeeper, app.poaKeeper)
	return []abci.ValidatorUpdate{}
}

// Initialize genesis.
func (app AppModule) InitGenesis(ctx sdk.Context, data json.RawMessage) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}

// Export genesis.
func (app AppModule) ExportGenesis(ctx sdk.Context) json.RawMessage {
	return json.RawMessage{}
}
