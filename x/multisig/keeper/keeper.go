package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/tendermint/tendermint/libs/log"
)

// Multisig keeper
type Keeper struct {
	storeKey sdk.StoreKey
	cdc      *codec.Codec
	router   Router
}

// Creating new multisig keeper
func NewKeeper(storeKey sdk.StoreKey, cdc *codec.Codec, router Router) Keeper {
	keeper := Keeper{
		storeKey: storeKey,
		cdc:      cdc,
		router:   router,
	}

	keeper.router.Seal()

	return keeper
}

// Get logger
func (keeper Keeper) getLogger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/multisig")
}

// Get keeper's codec
func (keeper Keeper) GetCDC() *codec.Codec {
	return keeper.cdc
}