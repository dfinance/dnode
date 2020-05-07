package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params/subspace"
	"github.com/tendermint/tendermint/libs/log"

	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/market/internal/types"
)

type Keeper struct {
	cdc           *codec.Codec
	storeKey      sdk.StoreKey
	paramSubspace subspace.Subspace
}

func NewKeeper(cdc *codec.Codec, paramStore subspace.Subspace) Keeper {
	return Keeper{
		cdc:           cdc,
		paramSubspace: paramStore.WithKeyTable(types.ParamKeyTable()),
	}
}

// Get logger for keeper.
func (k Keeper) GetLogger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/" + types.ModuleName)
}

func (k Keeper) nextID(params types.Params) dnTypes.ID {
	marketsLen := uint64(len(params.Markets))
	return dnTypes.NewIDFromUint64(marketsLen)
}

func (k Keeper) isNominee(ctx sdk.Context, nominee string) bool {
	params := k.GetParams(ctx)
	for _, v := range params.Nominees {
		if v == nominee {
			return true
		}
	}

	return false
}
