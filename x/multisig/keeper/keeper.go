package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"wings-blockchain/x/poa"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/gov"
)

// Multisig keeper
type Keeper struct {
	storeKey sdk.StoreKey
	poa		 poa.Keeper
	cdc      *codec.Codec
	router   gov.Router
}

// Creating new multisig keeper
func NewKeeper(storeKey sdk.StoreKey, poa poa.Keeper, cdc *codec.Codec, router gov.Router) Keeper {
	keeper := Keeper{
		storeKey: storeKey,
		poa:      poa,
		cdc:      cdc,
		router:   router,
	}

	keeper.router.Seal()
	return keeper
}