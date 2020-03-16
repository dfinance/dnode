package vmauth

import (
	"encoding/json"
	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authCli "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	authRoutes "github.com/cosmos/cosmos-sdk/x/auth/client/rest"
	authTypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	codec "github.com/tendermint/go-amino"
	abci "github.com/tendermint/tendermint/abci/types"
)

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

// app module basics object
type AppModuleBasic struct{}

// module name
func (AppModuleBasic) Name() string {
	return auth.ModuleName
}

// register module codec
func (AppModuleBasic) RegisterCodec(cdc *codec.Codec) {
	authTypes.RegisterCodec(cdc)
}

// default genesis state
func (AppModuleBasic) DefaultGenesis() json.RawMessage {
	return authTypes.ModuleCdc.MustMarshalJSON(authTypes.DefaultGenesisState())
}

// module validate genesis
func (AppModuleBasic) ValidateGenesis(bz json.RawMessage) error {
	var data authTypes.GenesisState
	if err := authTypes.ModuleCdc.UnmarshalJSON(bz, &data); err != nil {
		return err
	}

	return authTypes.ValidateGenesis(data)
}

// register rest routes
func (AppModuleBasic) RegisterRESTRoutes(ctx context.CLIContext, rtr *mux.Router) {
	authRoutes.RegisterRoutes(ctx, rtr, authTypes.StoreKey)
}

// get the root tx command of this module
func (AppModuleBasic) GetTxCmd(cdc *codec.Codec) *cobra.Command {
	return authCli.GetTxCmd(cdc)
}

// get the root query command of this module
func (AppModuleBasic) GetQueryCmd(cdc *codec.Codec) *cobra.Command {
	return authCli.GetQueryCmd(cdc)
}

//___________________________
// app module object
type AppModule struct {
	AppModuleBasic
	accountKeeper VMAccountKeeper
}

// NewAppModule creates a new AppModule object
func NewAppModule(accountKeeper VMAccountKeeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		accountKeeper:  accountKeeper,
	}
}

// module name
func (AppModule) Name() string {
	return authTypes.ModuleName
}

// register invariants
func (AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}

// module message route name
func (AppModule) Route() string { return "" }

// module handler
func (AppModule) NewHandler() sdk.Handler { return nil }

// module querier route name
func (AppModule) QuerierRoute() string {
	return authTypes.QuerierRoute
}

// module querier
func (am AppModule) NewQuerierHandler() sdk.Querier {
	return NewQuerier(am.accountKeeper)
}

// module init-genesis
func (am AppModule) InitGenesis(ctx sdk.Context, data json.RawMessage) []abci.ValidatorUpdate {
	var genesisState auth.GenesisState
	authTypes.ModuleCdc.MustUnmarshalJSON(data, &genesisState)
	am.accountKeeper.SetParams(ctx, genesisState.Params)

	return []abci.ValidatorUpdate{}
}

// module export genesis
func (am AppModule) ExportGenesis(ctx sdk.Context) json.RawMessage {
	params := am.accountKeeper.GetParams(ctx)

	return authTypes.ModuleCdc.MustMarshalJSON(auth.NewGenesisState(params))
}

// module begin-block
func (AppModule) BeginBlock(_ sdk.Context, _ abci.RequestBeginBlock) {}

// module end-block
func (AppModule) EndBlock(_ sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}
