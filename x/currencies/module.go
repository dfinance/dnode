package currencies

import (
	"encoding/json"
	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	codec "github.com/tendermint/go-amino"
	abci "github.com/tendermint/tendermint/abci/types"
	"wings-blockchain/x/core"
	"wings-blockchain/x/currencies/client"
	"wings-blockchain/x/currencies/client/rest"
	types "wings-blockchain/x/currencies/types"
)

var (
	_ core.AppMsModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

type AppModuleBasic struct{}

// Module name.
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// Registering codecs.
func (module AppModuleBasic) RegisterCodec(cdc *codec.Codec) {
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
func (AppModuleBasic) GetTxCmd(cdc *codec.Codec) *cobra.Command {
	return client.GetTxCmd(cdc)
}

// Get query commands for CLI.
func (AppModuleBasic) GetQueryCmd(cdc *codec.Codec) *cobra.Command {
	return client.GetQueryCmd(cdc)
}

// PoA module.
type AppModule struct {
	AppModuleBasic
	ccKeeper Keeper
}

// Create new PoA module.
func NewAppMsModule(ccKeeper Keeper) core.AppMsModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		ccKeeper:       ccKeeper,
	}
}

// Get name of module.
func (AppModule) Name() string {
	return types.ModuleName
}

// Register module invariants.
func (AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}

// Base route of module (for handler).
func (AppModule) Route() string { return types.Router }

// Create new handler.
func (app AppModule) NewHandler() sdk.Handler { return NewHandler(app.ccKeeper) }

// Create new multisignature handler.
func (app AppModule) NewMsHandler() core.MsHandler { return NewMsHandler(app.ccKeeper) }

// Get route for querier.
func (AppModule) QuerierRoute() string { return types.RouterKey }

// Get new querier for PoA module.
func (app AppModule) NewQuerierHandler() sdk.Querier {
	return NewQuerier(app.ccKeeper)
}

// Process begin block (abci).
func (AppModule) BeginBlock(_ sdk.Context, _ abci.RequestBeginBlock) {}

// Process end block (abci).
func (AppModule) EndBlock(_ sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
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
