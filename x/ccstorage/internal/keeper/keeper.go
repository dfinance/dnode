// Currencies storage module keeper stores currency and VM resources.
package keeper

import (
	cdcCodec "github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/dfinance/dnode/helpers/perms"
	"github.com/dfinance/dnode/x/ccstorage/internal/types"
	"github.com/dfinance/dnode/x/common_vm"
)

// Module keeper object.
type Keeper struct {
	cdc         *cdcCodec.Codec
	storeKey    sdk.StoreKey
	paramStore  params.Subspace
	vmKeeper    common_vm.VMStorage
	modulePerms perms.ModulePermissions
}

// GetLogger gets logger with keeper context.
func (k Keeper) GetLogger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// Create new currency storage keeper.
func NewKeeper(
	cdc *cdcCodec.Codec,
	storeKey sdk.StoreKey,
	paramSubspace params.Subspace,
	vmKeeper common_vm.VMStorage,
	permsRequesters ...perms.RequestModulePermissions,
) Keeper {
	k := Keeper{
		cdc:        cdc,
		storeKey:   storeKey,
		paramStore: paramSubspace.WithKeyTable(types.ParamKeyTable()),
		vmKeeper:   vmKeeper,
		modulePerms: types.NewModulePerms(),
	}
	for _, requester := range permsRequesters {
		k.modulePerms.AutoAddRequester(requester)
	}

	return k
}
