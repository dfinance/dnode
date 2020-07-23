// Module keeper used to integrate with other keepers and preserve clearanceState results.
package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/dfinance/dnode/helpers/perms"
	"github.com/dfinance/dnode/x/orderbook/internal/types"
	"github.com/dfinance/dnode/x/orders"
)

// Module keeper object.
type Keeper struct {
	cdc         *codec.Codec
	storeKey    sdk.StoreKey
	orderKeeper orders.Keeper
	modulePerms perms.ModulePermissions
}

// GetLogger gets logger with keeper context.
func (k Keeper) GetLogger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// GetOrderIterator returns orders module iterator (over orders).
func (k Keeper) GetOrderIterator(ctx sdk.Context) sdk.Iterator {
	k.modulePerms.AutoCheck(types.PermOrdersRead)

	return k.orderKeeper.GetIterator(ctx)
}

// ProcessOrderFills passes order fills to the orders module.
func (k Keeper) ProcessOrderFills(ctx sdk.Context, orderFills orders.OrderFills) {
	k.modulePerms.AutoCheck(types.PermExecFill)

	k.orderKeeper.ExecuteOrderFills(ctx, orderFills)
}

// NewKeeper creates keeper object.
func NewKeeper(
	cdc *codec.Codec,
	storeKey sdk.StoreKey,
	ok orders.Keeper,
	permsRequesters ...perms.RequestModulePermissions,
) Keeper {
	k := Keeper{
		cdc:         cdc,
		storeKey:    storeKey,
		orderKeeper: ok,
		modulePerms: types.NewModulePerms(),
	}
	for _, requester := range permsRequesters {
		k.modulePerms.AutoAddRequester(requester)
	}

	return k
}
