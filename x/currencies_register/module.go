package currencies_register

import (
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/dfinance/dnode/x/currencies_register/client"
	"github.com/dfinance/dnode/x/currencies_register/internal/keeper"
	"github.com/dfinance/dnode/x/currencies_register/internal/types"
)

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

type AppModuleBasic struct{}

// Module name.
func (AppModuleBasic) Name() string {
	return ModuleName
}

// Registering codecs (empty for now).
func (module AppModuleBasic) RegisterCodec(_ *codec.Codec) {}

// Validate exists genesis.
func (AppModuleBasic) ValidateGenesis(data json.RawMessage) error {
	var state types.GenesisState
	types.ModuleCdc.MustUnmarshalJSON(data, &state)

	denoms := make(map[string]bool)

	for _, genCurr := range state.Currencies {
		denom := genCurr.Denom
		if err := sdk.ValidateDenom(denom); err != nil {
			return fmt.Errorf("can't validate denom %q: %v", denom, err)
		}

		if denoms[denom] {
			return fmt.Errorf("doubled currency %q in genesis", denom)
		}

		denoms[denom] = true
	}

	return nil
}

// Generate default genesis.
func (module AppModuleBasic) DefaultGenesis() json.RawMessage {
	return types.ModuleCdc.MustMarshalJSON(types.DefaultGenesisState())
}

// Register REST routes.
func (AppModuleBasic) RegisterRESTRoutes(_ context.CLIContext, _ *mux.Router) {}

// Get transaction commands for CLI.
func (AppModuleBasic) GetTxCmd(_ *codec.Codec) *cobra.Command {
	return nil
}

// Get query commands for CLI.
func (AppModuleBasic) GetQueryCmd(cdc *codec.Codec) *cobra.Command {
	return client.GetQueryCmd(cdc)
}

// VM module.
type AppModule struct {
	AppModuleBasic
	keeper Keeper
}

// Create new VM module.
func NewAppModule(keeper Keeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		keeper:         keeper,
	}
}

// Get name of module.
func (AppModule) Name() string {
	return types.ModuleName
}

// Register module invariants.
func (AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}

// Base route of module (for handler).
func (AppModule) Route() string { return "" }

// Create new handler.
func (app AppModule) NewHandler() sdk.Handler {
	return nil
}

// Get route for querier.
func (AppModule) QuerierRoute() string {
	return RouterKey
}

// Get new querier for VM module.
func (app AppModule) NewQuerierHandler() sdk.Querier {
	return keeper.NewQuerier(app.keeper)
}

// Process begin block (abci).
func (AppModule) BeginBlock(_ sdk.Context, _ abci.RequestBeginBlock) {}

// Process end block (abci).
func (AppModule) EndBlock(_ sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}

// Initialize genesis.
func (app AppModule) InitGenesis(ctx sdk.Context, data json.RawMessage) []abci.ValidatorUpdate {
	err := app.keeper.InitGenesis(ctx, data)
	if err != nil {
		panic(err)
	}

	return []abci.ValidatorUpdate{}
}

// Export genesis.
func (app AppModule) ExportGenesis(ctx sdk.Context) json.RawMessage {
	return app.keeper.ExportGenesis(ctx)
}
