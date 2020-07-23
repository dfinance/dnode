// Multi signature module keeper stores call objects, calls queue with submitting, confirming and revoking.
package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/dfinance/dnode/helpers/perms"
	"github.com/dfinance/dnode/x/core/msmodule"
	"github.com/dfinance/dnode/x/multisig/internal/types"
)

// Module keeper object.
type Keeper struct {
	cdc         *codec.Codec
	storeKey    sdk.StoreKey
	paramStore  params.Subspace
	router      msmodule.MsRouter
	modulePerms perms.ModulePermissions
}

// GetLogger gets logger with keeper context.
func (k Keeper) GetLogger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// GetRouteHandler returns multi signature router handler for specific path.
func (k Keeper) GetRouteHandler(route string) msmodule.MsHandler {
	k.modulePerms.AutoCheck(types.PermRead)

	return k.router.GetRoute(route)
}

// Create new currency keeper.
func NewKeeper(
	cdc *codec.Codec,
	storeKey sdk.StoreKey,
	paramStore params.Subspace,
	router msmodule.MsRouter,
	permsRequesters ...perms.RequestModulePermissions,
) Keeper {
	k := Keeper{
		cdc:         cdc,
		storeKey:    storeKey,
		paramStore:  paramStore.WithKeyTable(types.ParamKeyTable()),
		router:      router,
		modulePerms: types.NewModulePerms(),
	}
	for _, requester := range permsRequesters {
		k.modulePerms.AutoAddRequester(requester)
	}

	return k
}
