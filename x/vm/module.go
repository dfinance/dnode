// VM module.
package vm

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/dfinance/dnode/x/vm/client/cli"
	types "github.com/dfinance/dnode/x/vm/internal/types"
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
func (module AppModuleBasic) RegisterCodec(cdc *codec.Codec) {
	types.RegisterCodec(cdc)
}

// Validate exists genesis.
func (AppModuleBasic) ValidateGenesis(data json.RawMessage) error {
	var state types.GenesisState
	types.ModuleCdc.MustUnmarshalJSON(data, &state)

	for _, genWriteOp := range state.WriteSet {
		bzAddr, err := hex.DecodeString(genWriteOp.Address)
		if err != nil {
			return err
		}

		// address length
		if len(bzAddr) != types.VmAddressLength {
			return fmt.Errorf("incorrect address %q length, should be %d bytes length", genWriteOp.Address, types.VmAddressLength)
		}

		if _, err := hex.DecodeString(genWriteOp.Path); err != nil {
			return err
		}

		if _, err := hex.DecodeString(genWriteOp.Value); err != nil {
			return err
		}
	}

	return nil
}

// Generate default genesis.
func (module AppModuleBasic) DefaultGenesis() json.RawMessage {
	return types.ModuleCdc.MustMarshalJSON(&types.GenesisState{})
}

// Register REST routes.
func (AppModuleBasic) RegisterRESTRoutes(ctx context.CLIContext, r *mux.Router) {}

// Get transaction commands for CLI.
func (AppModuleBasic) GetTxCmd(cdc *codec.Codec) *cobra.Command {
	return cli.GetTxCmd(cdc)
}

// Get query commands for CLI.
func (AppModuleBasic) GetQueryCmd(cdc *codec.Codec) *cobra.Command {
	return cli.GetQueriesCmd(cdc)
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
	return NewQuerier(app.vmKeeper)
}

// Process begin block (abci).
func (AppModule) BeginBlock(_ sdk.Context, _ abci.RequestBeginBlock) {}

// Process end block (abci).
func (AppModule) EndBlock(_ sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}

// Initialize genesis.
func (app AppModule) InitGenesis(ctx sdk.Context, data json.RawMessage) []abci.ValidatorUpdate {
	app.vmKeeper.InitGenesis(ctx, data)
	return []abci.ValidatorUpdate{}
}

// Export genesis.
func (app AppModule) ExportGenesis(ctx sdk.Context) json.RawMessage {
	genesisState := app.vmKeeper.ExportGenesis(ctx)
	return types.ModuleCdc.MustMarshalJSON(genesisState)
}
