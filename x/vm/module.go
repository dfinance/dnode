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
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/dfinance/dnode/x/common_vm"
	"github.com/dfinance/dnode/x/vm/client"
	"github.com/dfinance/dnode/x/vm/client/rest"
	"github.com/dfinance/dnode/x/vm/internal/keeper"
)

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

type AppModuleBasic struct{}

// Module name.
func (module AppModuleBasic) Name() string {
	return ModuleName
}

// Registering codecs.
func (module AppModuleBasic) RegisterCodec(cdc *codec.Codec) {
	RegisterCodec(cdc)
}

// Validate exists genesis.
func (module AppModuleBasic) ValidateGenesis(data json.RawMessage) error {
	var state GenesisState
	ModuleCdc.MustUnmarshalJSON(data, &state)

	for _, genWriteOp := range state.WriteSet {
		bzAddr, err := hex.DecodeString(genWriteOp.Address)
		if err != nil {
			return err
		}

		// address length
		if len(bzAddr) != common_vm.VMAddressLength {
			return fmt.Errorf("incorrect address %q length, should be %d bytes length", genWriteOp.Address, common_vm.VMAddressLength)
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
	return ModuleCdc.MustMarshalJSON(&GenesisState{})
}

// Register REST routes.
func (module AppModuleBasic) RegisterRESTRoutes(ctx context.CLIContext, r *mux.Router) {
	rest.RegisterRoutes(ctx, r)
}

// Get transaction commands for CLI.
func (module AppModuleBasic) GetTxCmd(cdc *codec.Codec) *cobra.Command {
	return client.GetTxCmd(cdc)
}

// Get query commands for CLI.
func (module AppModuleBasic) GetQueryCmd(cdc *codec.Codec) *cobra.Command {
	return client.GetQueryCmd(cdc)
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
func (app AppModule) Name() string {
	return ModuleName
}

// Register module invariants.
func (app AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}

// Base route of module (for handler).
func (app AppModule) Route() string { return RouterKey }

// Create new handler.
func (app AppModule) NewHandler() sdk.Handler { return NewHandler(app.vmKeeper) }

// Create governance handler.
func (app AppModule) NewGovHandler() gov.Handler { return NewGovHandler(app.vmKeeper) }

// Get route for querier.
func (app AppModule) QuerierRoute() string { return RouterKey }

// Get new querier for VM module.
func (app AppModule) NewQuerierHandler() sdk.Querier {
	return keeper.NewQuerier(app.vmKeeper)
}

// Process begin block (abci).
func (app AppModule) BeginBlock(ctx sdk.Context, req abci.RequestBeginBlock) {
	BeginBlocker(ctx, app.vmKeeper, req)
}

// Process end block (abci).
func (app AppModule) EndBlock(ctx sdk.Context, _ abci.RequestEndBlock) []abci.ValidatorUpdate {
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
	return ModuleCdc.MustMarshalJSON(genesisState)
}
