// Keeper implementation.
package multisig

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/WingsDao/wings-blockchain/x/core"
)

// Multisignature keeper.
type Keeper struct {
	storeKey sdk.StoreKey
	cdc      *codec.Codec
	router   core.Router
}

// Creating new multisignature keeper implementation.
func NewKeeper(storeKey sdk.StoreKey, cdc *codec.Codec, router core.Router) Keeper {
	keeper := Keeper{
		storeKey: storeKey,
		cdc:      cdc,
		router:   router,
	}

	return keeper
}

// Get logger for keeper.
func (keeper Keeper) getLogger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/multisig")
}
