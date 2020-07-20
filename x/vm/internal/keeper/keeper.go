// VM keeper processing messages from handler.
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

// VM keeper.
type Keeper struct {
	cdc      *amino.Codec // Amino codec.
	storeKey sdk.StoreKey // Store key.

	client    VMClient         // VM service client.
	listener  net.Listener     // VM data server listener.
	rawClient *grpc.ClientConn // GRPC connection to VM.

	config *config.VMConfig // VM config.

	dsServer    *DSServer    // Data-source server.
	rawDSServer *grpc.Server // GRPC raw server.

	modulePerms perms.ModulePermissions
}

// Check that VMStorage is compatible with keeper (later we can do it by events probably).
var _ common_vm.VMStorage = Keeper{}

// VM keeper logger.
func (Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// Execute script.
func (keeper Keeper) ExecuteScript(ctx sdk.Context, msg types.MsgExecuteScript) error {
	keeper.modulePerms.AutoCheck(types.PermVmExec)

	req, sdkErr := NewExecuteRequest(ctx, msg)
	if sdkErr != nil {
		return sdkErr
	}

	exec, err := keeper.sendExecuteReq(ctx, nil, req)
	if err != nil {
		keeper.Logger(ctx).Error(fmt.Sprintf("grpc error: %s", err.Error()))
		panic(sdkErrors.Wrap(types.ErrVMCrashed, err.Error()))
	}

	keeper.processExecution(ctx, exec)

	return nil
}

// Execute script without response processing (used for debug).
func (keeper Keeper) ExecuteScriptNoProcessing(ctx sdk.Context, msg types.MsgExecuteScript) (*vm_grpc.VMExecuteResponse, error) {
	keeper.modulePerms.AutoCheck(types.PermVmExec)

	req, sdkErr := NewExecuteRequest(ctx, msg)
	if sdkErr != nil {
		return nil, sdkErr
	}

	exec, err := keeper.sendExecuteReq(ctx, nil, req)
	if err != nil {
		return nil, err
	}

	return exec, nil
}

// Deploy module.
func (keeper Keeper) DeployContract(ctx sdk.Context, msg types.MsgDeployModule) error {
	keeper.modulePerms.AutoCheck(types.PermVmExec)

	req, sdkErr := NewDeployRequest(ctx, msg)
	if sdkErr != nil {
		return sdkErr
	}

	exec, err := keeper.sendExecuteReq(ctx, req, nil)
	if err != nil {
		keeper.Logger(ctx).Error(fmt.Sprintf("grpc error: %s", err.Error()))
		panic(sdkErrors.Wrap(types.ErrVMCrashed, err.Error()))
	}

	keeper.processExecution(ctx, exec)

	return nil
}

// DeployContractDryRun checks that contract can be deployed (returned writeSets are not persisted to store).
func (keeper Keeper) DeployContractDryRun(ctx sdk.Context, msg types.MsgDeployModule) error {
	keeper.modulePerms.AutoCheck(types.PermVmExec)

	req, sdkErr := NewDeployRequest(ctx, msg)
	if sdkErr != nil {
		return sdkErr
	}

	exec, dvmErr := keeper.sendExecuteReq(ctx, req, nil)
	if dvmErr != nil {
		return sdkErrors.Wrap(types.ErrVMCrashed, dvmErr.Error())
	}

	if exec.Status != vm_grpc.ContractStatus_Discard {
		if exec.StatusStruct != nil && exec.StatusStruct.MajorStatus != types.VMCodeExecuted {
			statusMsg := types.VMExecStatusToString(exec.Status, exec.StatusStruct)
			return sdkErrors.Wrap(types.ErrWrongExecutionResponse, statusMsg)
		}
	}

	return nil
}

// Initialize VM keeper (include grpc client to VM and grpc server for data store).
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
