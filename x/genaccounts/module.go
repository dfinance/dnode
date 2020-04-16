package genaccounts

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/dfinance/dnode/x/genaccounts/internal/types"
)

var (
	_ module.AppModuleGenesis = AppModule{}
	_ module.AppModuleBasic   = AppModuleBasic{}
)

// App module basics object.
type AppModuleBasic struct{}

// Module name.
func (AppModuleBasic) Name() string {
	return ModuleName
}

// Register module codec.
func (AppModuleBasic) RegisterCodec(cdc *codec.Codec) {}

// Default genesis state.
func (AppModuleBasic) DefaultGenesis() json.RawMessage {
	return ModuleCdc.MustMarshalJSON(GenesisState{})
}

// Module validate genesis.
func (AppModuleBasic) ValidateGenesis(bz json.RawMessage) error {
	var data GenesisState
	err := ModuleCdc.UnmarshalJSON(bz, &data)
	if err != nil {
		return err
	}

	return ValidateGenesis(data)
}

// Register rest routes.
func (AppModuleBasic) RegisterRESTRoutes(_ context.CLIContext, _ *mux.Router) {}

// Get the root tx command of this module.
func (AppModuleBasic) GetTxCmd(_ *codec.Codec) *cobra.Command { return nil }

// Get the root query command of this module.
func (AppModuleBasic) GetQueryCmd(_ *codec.Codec) *cobra.Command { return nil }

// Iterate the genesis accounts and perform an operation at each of them.
// - to used by other modules
func (AppModuleBasic) IterateGenesisAccounts(cdc *codec.Codec, appGenesis map[string]json.RawMessage,
	iterateFn func(exported.Account) (stop bool)) {

	genesisState := GetGenesisStateFromAppState(cdc, appGenesis)
	for _, ga := range genesisState {
		acc := ga.ToAccount()
		if iterateFn(acc) {
			break
		}
	}
}

type AppModule struct {
	AppModuleBasic
	accountKeeper types.AccountKeeper
}

// NewAppModule creates a new AppModule object.
func NewAppModule(accountKeeper types.AccountKeeper) module.AppModule {
	return module.NewGenesisOnlyAppModule(AppModule{
		AppModuleBasic: AppModuleBasic{},
		accountKeeper:  accountKeeper,
	})
}

// module init genesis.
func (am AppModule) InitGenesis(ctx sdk.Context, data json.RawMessage) []abci.ValidatorUpdate {
	genesisState := GenesisState{}
	ModuleCdc.MustUnmarshalJSON(data, &genesisState)

	InitGenesis(ctx, ModuleCdc, am.accountKeeper, genesisState)

	return []abci.ValidatorUpdate{}
}

// Module export genesis.
func (am AppModule) ExportGenesis(ctx sdk.Context) json.RawMessage {
	genesisState := ExportGenesis(ctx, am.accountKeeper)

	return ModuleCdc.MustMarshalJSON(genesisState)
}
