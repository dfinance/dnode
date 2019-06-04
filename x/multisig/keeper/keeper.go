package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/codec"
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