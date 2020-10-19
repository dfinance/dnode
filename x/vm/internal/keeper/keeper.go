// VM module keeper wraps DVM gRPC DataServer service and gRPC VM service client.
// Keeper provides VM storage functionality used by other modules.
// DataSource server provides async data (writeSets) requests from DVM to VM during VM script/module execution.
package keeper

import (
	"fmt"
	"net"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/tendermint/go-amino"
	"github.com/tendermint/tendermint/libs/log"
	"google.golang.org/grpc"

	"github.com/dfinance/dvm-proto/go/vm_grpc"

	"github.com/dfinance/dnode/cmd/config"
	"github.com/dfinance/dnode/helpers/perms"
	"github.com/dfinance/dnode/x/common_vm"
	"github.com/dfinance/dnode/x/vm/internal/middlewares"
	"github.com/dfinance/dnode/x/vm/internal/types"
)

// Check that VMStorage is compatible with keeper (later we can do it by events probably).
var _ common_vm.VMStorage = Keeper{}

// Module keeper object.
type Keeper struct {
	cdc      *amino.Codec
	storeKey sdk.StoreKey
	//
	config *config.VMConfig
	// VM connection
	client    VMClient         // aggregated gRPC services client
	rawClient *grpc.ClientConn // gRPC connection
	// DataSource server
	listener    net.Listener
	dsServer    *DSServer
	rawDSServer *grpc.Server
	//
	modulePerms perms.ModulePermissions
}

// GetLogger gets logger with keeper context.
func (k Keeper) GetLogger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// ExecuteScript executes Move script and processes execution results (events, writeSets).
func (k Keeper) ExecuteScript(ctx sdk.Context, msg types.MsgExecuteScript) error {
	k.modulePerms.AutoCheck(types.PermVmExec)

	req, sdkErr := NewExecuteRequest(ctx, msg)
	if sdkErr != nil {
		return sdkErr
	}

	exec, err := k.sendExecuteReq(ctx, nil, req)
	if err != nil {
		k.GetLogger(ctx).Error(fmt.Sprintf("grpc error: %s", err.Error()))
		panic(sdkErrors.Wrap(types.ErrVMCrashed, err.Error()))
	}

	k.processExecution(ctx, exec)

	return nil
}

// ExecuteScriptNoProcessing is executes Move script without execution processing (used for debug).
func (k Keeper) ExecuteScriptNoProcessing(ctx sdk.Context, msg types.MsgExecuteScript) (*vm_grpc.VMExecuteResponse, error) {
	k.modulePerms.AutoCheck(types.PermVmExec)

	req, sdkErr := NewExecuteRequest(ctx, msg)
	if sdkErr != nil {
		return nil, sdkErr
	}

	exec, err := k.sendExecuteReq(ctx, nil, req)
	if err != nil {
		return nil, err
	}

	return exec, nil
}

// DeployContract deploys Move module (contract) and processes execution results (events, writeSets).
func (k Keeper) DeployContract(ctx sdk.Context, msg types.MsgDeployModule) error {
	k.modulePerms.AutoCheck(types.PermVmExec)

	execList := make([]*vm_grpc.VMExecuteResponse, len(msg.Module))
	var intErr error

	for i, contact := range msg.Module {
		req, sdkErr := NewDeployRequest(ctx, msg.Signer, contact)
		if sdkErr != nil {
			return sdkErr
		}

		execList[i], intErr = k.sendExecuteReq(ctx, req, nil)
		if intErr != nil {
			k.GetLogger(ctx).Error(fmt.Sprintf("grpc error: %s", intErr.Error()))
			panic(sdkErrors.Wrap(types.ErrVMCrashed, intErr.Error()))
		}
	}

	for _, exec := range execList {
		k.processExecution(ctx, exec)
	}

	return nil
}

// DeployContractDryRun checks that contract can be deployed (returned writeSets are not persisted to store).
func (k Keeper) DeployContractDryRun(ctx sdk.Context, msg types.MsgDeployModule) error {
	k.modulePerms.AutoCheck(types.PermVmExec)

	for _, contact := range msg.Module {
		req, sdkErr := NewDeployRequest(ctx, msg.Signer, contact)
		if sdkErr != nil {
			return sdkErr
		}

		exec, dvmErr := k.sendExecuteReq(ctx, req, nil)
		if dvmErr != nil {
			return sdkErrors.Wrap(types.ErrVMCrashed, dvmErr.Error())
		}

		if exec.GetStatus().GetError() != nil {
			statusMsg := types.StringifyVMExecStatus(exec.Status)
			return sdkErrors.Wrap(types.ErrWrongExecutionResponse, statusMsg)
		}
	}

	return nil
}

// Create new currency keeper.
func NewKeeper(
	cdc *amino.Codec,
	storeKey sdk.StoreKey,
	conn *grpc.ClientConn,
	listener net.Listener,
	config *config.VMConfig,
	permsRequesters ...perms.RequestModulePermissions,
) Keeper {
	keeper := Keeper{
		cdc:         cdc,
		storeKey:    storeKey,
		rawClient:   conn,
		client:      NewVMClient(conn),
		listener:    listener,
		config:      config,
		modulePerms: types.NewModulePerms(),
	}
	for _, requester := range permsRequesters {
		keeper.modulePerms.AutoAddRequester(requester)
	}

	keeper.dsServer = NewDSServer(&keeper)
	keeper.dsServer.RegisterDataMiddleware(middlewares.NewBlockMiddleware())
	keeper.dsServer.RegisterDataMiddleware(middlewares.NewTimeMiddleware())

	return keeper
}
