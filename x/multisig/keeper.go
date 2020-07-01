// Keeper implementation.
package multisig

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/dfinance/dnode/x/core/msmodule"
)

// Multisignature keeper.
type Keeper struct {
	storeKey   sdk.StoreKey
	cdc        *codec.Codec
	router     msmodule.Router
	paramStore params.Subspace
}

// Creating new multisignature keeper implementation.
func NewKeeper(storeKey sdk.StoreKey, cdc *codec.Codec, router msmodule.Router, paramStore params.Subspace) Keeper {
	keeper := Keeper{
		storeKey:   storeKey,
		cdc:        cdc,
		router:     router,
		paramStore: paramStore.WithKeyTable(NewKeyTable()),
	}

	return keeper
}

// Get logger for keeper.
func (keeper Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", ModuleName))
}
