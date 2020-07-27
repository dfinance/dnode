package fake

import (
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	"github.com/tendermint/go-amino"
	abci "github.com/tendermint/tendermint/abci/types"
)

const (
	ModuleName = "fake"
)

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
	//
	ModuleCdc = codec.New()
	genesis   GenesisState
)

type GenesisState struct {
	StrValue string `json:"str_value" yaml:"str_value"`
	IntValue int    `json:"int_value" yaml:"int_value"`
}

func DefaultGenesis() GenesisState {
	return GenesisState{
		StrValue: "genesis_string",
		IntValue: 42,
	}
}

// AppModuleBasic app module basics object.
type AppModuleBasic struct{}

// Name gets module name.
func (AppModuleBasic) Name() string { return ModuleName }

// RegisterCodec registers module codec.
func (AppModuleBasic) RegisterCodec(cdc *codec.Codec) {}

// DefaultGenesis gets default module genesis state.
func (AppModuleBasic) DefaultGenesis() json.RawMessage {
	return ModuleCdc.MustMarshalJSON(DefaultGenesis)
}

// ValidateGenesis validates module genesis state.
func (AppModuleBasic) ValidateGenesis(bz json.RawMessage) error {
	var state GenesisState
	ModuleCdc.MustUnmarshalJSON(bz, &state)

	if state.IntValue != 42 {
		return fmt.Errorf("int_value: invalid")
	}

	return nil
}

// RegisterRESTRoutes registers module REST routes.
func (AppModuleBasic) RegisterRESTRoutes(ctx context.CLIContext, r *mux.Router) {}

// GetTxCmd returns module root tx command.
func (AppModuleBasic) GetTxCmd(cdc *amino.Codec) *cobra.Command { return nil }

// GetQueryCmd returns module root query command.
func (AppModuleBasic) GetQueryCmd(cdc *amino.Codec) *cobra.Command { return nil }

// AppModule is a app module type.
type AppModule struct {
	AppModuleBasic
}

// NewAppModule creates new AppModule object.
func NewAppModule() AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{},
	}
}

// Name gets module name.
func (app AppModule) Name() string { return ModuleName }

// RegisterInvariants registers module invariants.
func (app AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}

// Route returns module messages route.
func (app AppModule) Route() string { return "" }

// NewHandler returns module messages handler.
func (app AppModule) NewHandler() sdk.Handler { return nil }

// QuerierRoute returns module querier route.
func (app AppModule) QuerierRoute() string { return "" }

// NewQuerierHandler creates module querier.
func (app AppModule) NewQuerierHandler() sdk.Querier { return nil }

// InitGenesis inits module-genesis state.
func (app AppModule) InitGenesis(ctx sdk.Context, data json.RawMessage) []abci.ValidatorUpdate {
	var state GenesisState
	ModuleCdc.MustUnmarshalJSON(data, &state)
	genesis = state

	return []abci.ValidatorUpdate{}
}

// ExportGenesis exports module genesis state.
func (app AppModule) ExportGenesis(ctx sdk.Context) json.RawMessage {
	return ModuleCdc.MustMarshalJSON(genesis)
}

// BeginBlock performs module actions at a block start.
func (app AppModule) BeginBlock(_ sdk.Context, _ abci.RequestBeginBlock) {}

// EndBlock performs module actions at a block end.
func (app AppModule) EndBlock(ctx sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}

func init() {
	codec.RegisterCrypto(ModuleCdc)
}
