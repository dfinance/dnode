//
//
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
	msTypes "wings-blockchain/x/multisig/types"
	"wings-blockchain/x/poa/client"
	"wings-blockchain/x/poa/client/rest"
	"wings-blockchain/x/poa/types"
)

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

type AppModuleBasic struct{}

func (AppModuleBasic) Name() string {
	return types.ModuleName
}

func (module AppModuleBasic) RegisterCodec(cdc *codec.Codec) {
	RegisterCodec(cdc)
}

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

func (module AppModuleBasic) DefaultGenesis() json.RawMessage {
	return ModuleCdc.MustMarshalJSON(types.GenesisState{
		Parameters:    types.DefaultParams(),
		PoAValidators: types.Validators{},
	})
}

func (AppModuleBasic) RegisterRESTRoutes(ctx context.CLIContext, r *mux.Router) {
	rest.RegisterRoutes(ctx, r)
}

func (AppModuleBasic) GetTxCmd(cdc *codec.Codec) *cobra.Command {
	return client.GetTxCmd(cdc)
}

func (AppModuleBasic) GetQueryCmd(cdc *codec.Codec) *cobra.Command {
	return client.GetQueryCmd(cdc)
}

type AppModule struct {
	AppModuleBasic
	poaKeeper Keeper
}

func NewAppModule(poaKeeper Keeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
		poaKeeper:      poaKeeper,
	}
}

func (AppModule) Name() string {
	return types.ModuleName
}

func (AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}

func (AppModule) Route() string { return types.RouterKey }

func (app AppModule) NewHandler() sdk.Handler { return NewHandler(app.poaKeeper) }

func (app AppModule) NewMsHandler() msTypes.MsHandler { return NewMsHandler(app.poaKeeper) }

func (AppModule) QuerierRoute() string { return types.RouterKey }

func (app AppModule) NewQuerierHandler() sdk.Querier {
	return NewQuerier(app.poaKeeper)
}

func (AppModule) BeginBlock(_ sdk.Context, _ abci.RequestBeginBlock) {}

func (AppModule) EndBlock(_ sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}

func (app AppModule) InitGenesis(ctx sdk.Context, data json.RawMessage) []abci.ValidatorUpdate {
	var genesisState types.GenesisState

	ModuleCdc.MustUnmarshalJSON(data, &genesisState)

	app.poaKeeper.InitGenesis(ctx, genesisState)

	return []abci.ValidatorUpdate{}
}

func (app AppModule) ExportGenesis(ctx sdk.Context) json.RawMessage {
	genesisState := app.poaKeeper.ExportGenesis(ctx)
	return ModuleCdc.MustMarshalJSON(genesisState)
}
