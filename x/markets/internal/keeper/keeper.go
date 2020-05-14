// Markets module keeper creates and stores markets objects.
package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params/subspace"
	"github.com/tendermint/tendermint/libs/log"

	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/currencies_register"
	"github.com/dfinance/dnode/x/markets/internal/types"
)

// Module keeper object.
type Keeper struct {
	cdc              *codec.Codec
	storeKey         sdk.StoreKey
	paramSubspace    subspace.Subspace
	ccRegisterKeeper currencies_register.Keeper
}

// NewKeeper creates keeper object.
func NewKeeper(cdc *codec.Codec, paramStore subspace.Subspace, crk currencies_register.Keeper) Keeper {
	return Keeper{
		cdc:              cdc,
		paramSubspace:    paramStore.WithKeyTable(types.ParamKeyTable()),
		ccRegisterKeeper: crk,
	}
}

// GetLogger gets logger with keeper context.
func (k Keeper) GetLogger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// nextID return next unique market object ID.
func (k Keeper) nextID(params types.Params) dnTypes.ID {
	marketsLen := uint64(len(params.Markets))
	return dnTypes.NewIDFromUint64(marketsLen)
}
