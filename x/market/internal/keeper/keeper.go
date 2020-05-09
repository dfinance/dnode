// Market module keeper creates and stores market objects.
// Market objects creation is only allowed for nominee accounts.
package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params/subspace"
	"github.com/tendermint/tendermint/libs/log"

	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/market/internal/types"
)

// Module keeper object.
type Keeper struct {
	cdc           *codec.Codec
	storeKey      sdk.StoreKey
	paramSubspace subspace.Subspace
}

// NewKeeper creates keeper object.
func NewKeeper(cdc *codec.Codec, paramStore subspace.Subspace) Keeper {
	return Keeper{
		cdc:           cdc,
		paramSubspace: paramStore.WithKeyTable(types.ParamKeyTable()),
	}
}

// GetLogger gets logger with keeper context.
func (k Keeper) GetLogger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/" + types.ModuleName)
}

// nextID return next unique market object ID.
func (k Keeper) nextID(params types.Params) dnTypes.ID {
	marketsLen := uint64(len(params.Markets))
	return dnTypes.NewIDFromUint64(marketsLen)
}

// isNominee checks if account as a nominee account.
func (k Keeper) isNominee(ctx sdk.Context, nominee string) bool {
	params := k.GetParams(ctx)
	for _, v := range params.Nominees {
		if v == nominee {
			return true
		}
	}

	return false
}
