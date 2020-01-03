package multisig

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"
	"wings-blockchain/x/core"
)

// Multisig keeper
type Keeper struct {
	storeKey sdk.StoreKey
	cdc      *codec.Codec
	router   core.Router
}

// Creating new multisig keeper
func NewKeeper(storeKey sdk.StoreKey, cdc *codec.Codec, router core.Router) Keeper {
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
