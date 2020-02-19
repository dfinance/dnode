// Implements PoA module.
package poa

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/WingsDao/wings-blockchain/x/core"
	"github.com/WingsDao/wings-blockchain/x/poa/client"
	"github.com/WingsDao/wings-blockchain/x/poa/client/rest"
	"github.com/WingsDao/wings-blockchain/x/poa/types"
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
	var genesisState types.GenesisState
	err := ModuleCdc.UnmarshalJSON(bz, &genesisState)
	if err != nil {
		return err
	}

	params := genesisState.Parameters
	if err = params.Validate(); err != nil {
		return err
	}

	length := len(genesisState.PoAValidators)

	if length < int(params.MinValidators) {
		return types.ErrNotEnoungValidators(uint16(length), params.MinValidators)
	}

	if length > int(params.MaxValidators) {
		return types.ErrMaxValidatorsReached(params.MaxValidators)
	}

	return nil
}

// Generate default genesis.
func (module AppModuleBasic) DefaultGenesis() json.RawMessage {
	return ModuleCdc.MustMarshalJSON(types.GenesisState{
		Parameters:    types.DefaultParams(),
		PoAValidators: types.Validators{},
	})
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
	poaKeeper Keeper
}

// Create new PoA module.
func NewAppMsModule(poaKeeper Keeper) core.AppMsModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
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
func (app AppModule) NewHandler() sdk.Handler { return NewHandler(app.poaKeeper) }

// Create new multisignature handler.
func (app AppModule) NewMsHandler() core.MsHandler { return NewMsHandler(app.poaKeeper) }

// Get route for querier.
func (AppModule) QuerierRoute() string { return types.RouterKey }

// Get new querier for PoA module.
func (app AppModule) NewQuerierHandler() sdk.Querier {
	return NewQuerier(app.poaKeeper)
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

	ModuleCdc.MustUnmarshalJSON(data, &genesisState)

	app.poaKeeper.InitGenesis(ctx, genesisState)

	return []abci.ValidatorUpdate{}
}

// Export genesis.
func (app AppModule) ExportGenesis(ctx sdk.Context) json.RawMessage {
	genesisState := app.poaKeeper.ExportGenesis(ctx)
	return ModuleCdc.MustMarshalJSON(genesisState)
}
