package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"

	orderTypes "github.com/dfinance/dnode/x/order"
	"github.com/dfinance/dnode/x/orderbook/internal/types"
)

type Keeper struct {
	cdc         *codec.Codec
	orderKeeper orderTypes.Keeper
}

func NewKeeper(cdc *codec.Codec, ok orderTypes.Keeper) Keeper {
	return Keeper{
		cdc:         cdc,
		orderKeeper: ok,
	}
}

// Get logger for keeper.
func (k Keeper) GetLogger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/" + types.ModuleName)
}

func (k Keeper) GetOrderIterator(ctx sdk.Context) sdk.Iterator {
	return k.orderKeeper.GetIterator(ctx)
}

func (k Keeper) ProcessOrderFills(ctx sdk.Context, orderFills orderTypes.OrderFills) {
	k.orderKeeper.ExecuteOrderFills(ctx, orderFills)
}
