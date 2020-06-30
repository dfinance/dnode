// Currencies module registers / issues and withdraws currencies.
// Module is integrated with VM for CurrencyInfo and Balance resources.
// Issue is a multisig operation.
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

	"github.com/dfinance/dnode/x/core"
	"github.com/dfinance/dnode/x/currencies/client"
	"github.com/dfinance/dnode/x/currencies/client/rest"
	"github.com/dfinance/dnode/x/currencies/internal/keeper"
	types2 "github.com/dfinance/dnode/x/currencies/internal/types"
)

var (
	_ core.AppMsModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

// AppModuleBasic app module basics object.
type AppModuleBasic struct{}

// Name gets module name.
func (AppModuleBasic) Name() string {
	return types2.ModuleName
}

// RegisterCodec registers module codec.
func (AppModuleBasic) RegisterCodec(cdc *codec.Codec) {
	types2.RegisterCodec(cdc)
}

// DefaultGenesis gets default module genesis state.
func (AppModuleBasic) DefaultGenesis() json.RawMessage {
	return ModuleCdc.MustMarshalJSON(DefaultGenesisState())
}

// ValidateGenesis validates module genesis state.
func (AppModuleBasic) ValidateGenesis(bz json.RawMessage) error {
	state := GenesisState{}
	ModuleCdc.MustUnmarshalJSON(bz, &state)

	return state.Validate()
}

// RegisterRESTRoutes registers module REST routes.
func (AppModuleBasic) RegisterRESTRoutes(ctx context.CLIContext, r *mux.Router) {
	rest.RegisterRoutes(ctx, r)
}

// GetTxCmd returns module root tx command.
func (AppModuleBasic) GetTxCmd(cdc *codec.Codec) *cobra.Command {
	return client.GetTxCmd(cdc)
}

// GetQueryCmd returns module root query command.
func (AppModuleBasic) GetQueryCmd(cdc *codec.Codec) *cobra.Command {
	return client.GetQueryCmd(cdc)
}

// AppModule is a app module type.
type AppModule struct {
	AppModuleBasic
	ccKeeper keeper.Keeper
}

// NewAppMsModule creates new AppMsModule object.
func NewAppMsModule(ccKeeper keeper.Keeper) core.AppMsModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		ccKeeper:       ccKeeper,
	}
}

// Name gets module name.
func (AppModule) Name() string {
	return types2.ModuleName
}

// RegisterInvariants registers module invariants.
func (app AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}

// Route returns module messages route.
func (app AppModule) Route() string {
	return RouterKey
}

// NewHandler returns module messages handler.
func (app AppModule) NewHandler() sdk.Handler {
	return NewHandler(app.ccKeeper)
}

// NewMsHandler returns module multisig messages handler.
func (app AppModule) NewMsHandler() core.MsHandler {
	return NewMsHandler(app.ccKeeper)
}

// QuerierRoute returns module querier route.
func (app AppModule) QuerierRoute() string {
	return types2.RouterKey
}

// NewQuerierHandler creates module querier.
func (app AppModule) NewQuerierHandler() sdk.Querier {
	return keeper.NewQuerier(app.ccKeeper)
}

// InitGenesis inits module-genesis state.
func (app AppModule) InitGenesis(ctx sdk.Context, data json.RawMessage) []abci.ValidatorUpdate {
	app.ccKeeper.InitGenesis(ctx, data)

	return []abci.ValidatorUpdate{}
}

// ExportGenesis exports module genesis state.
func (app AppModule) ExportGenesis(ctx sdk.Context) json.RawMessage {
	return app.ccKeeper.ExportGenesis(ctx)
}

// BeginBlock performs module actions at a block start.
func (app AppModule) BeginBlock(_ sdk.Context, _ abci.RequestBeginBlock) {}

// EndBlock performs module actions at a block end.
func (app AppModule) EndBlock(_ sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}
