// Module keeper used to integrate with other keepers.
package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/dfinance/dnode/x/orderbook/internal/types"
	orderTypes "github.com/dfinance/dnode/x/orders"
)

// Module keeper object.
type Keeper struct {
	cdc         *codec.Codec
	orderKeeper orderTypes.Keeper
}

// NewKeeper creates keeper object.
func NewKeeper(cdc *codec.Codec, ok orderTypes.Keeper) Keeper {
	return Keeper{
		cdc:         cdc,
		orderKeeper: ok,
	}
}

// GetLogger gets logger with keeper context.
func (k Keeper) GetLogger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// GetOrderIterator returns orders module iterator (over orders).
func (k Keeper) GetOrderIterator(ctx sdk.Context) sdk.Iterator {
	return k.orderKeeper.GetIterator(ctx)
}

// ProcessOrderFills passes order fills to the orders module.
func (k Keeper) ProcessOrderFills(ctx sdk.Context, orderFills orderTypes.OrderFills) {
	k.orderKeeper.ExecuteOrderFills(ctx, orderFills)
}
