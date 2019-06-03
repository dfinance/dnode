package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"wings-blockchain/x/poa"
	"github.com/cosmos/cosmos-sdk/codec"
	"wings-blockchain/x/multisig"
)

// Multisig keeper
type Keeper struct {
	storeKey sdk.StoreKey
	poa		 poa.Keeper
	cdc      *codec.Codec
	router   multisig.Router
}

// Creating new multisig keeper
func NewKeeper(storeKey sdk.StoreKey, cdc *codec.Codec, router multisig.Router) Keeper {
	keeper := Keeper{
		storeKey: storeKey,
		cdc:      cdc,
		router:   router,
	}

	keeper.router.Seal()
	return keeper
}