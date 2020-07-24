// Markets module keeper creates and stores markets objects.
package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params/subspace"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/dfinance/dnode/helpers/perms"
	"github.com/dfinance/dnode/x/ccstorage"
	"github.com/dfinance/dnode/x/markets/internal/types"
)

// Module keeper object.
type Keeper struct {
	cdc           *codec.Codec
	storeKey      sdk.StoreKey
	paramSubspace subspace.Subspace
	ccsStorage    ccstorage.Keeper
	modulePerms   perms.ModulePermissions
}

// GetLogger gets logger with keeper context.
func (k Keeper) GetLogger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// NewKeeper creates keeper object.
func NewKeeper(
	cdc *codec.Codec,
	storeKey sdk.StoreKey,
	ccsKeeper ccstorage.Keeper,
	permsRequesters ...perms.RequestModulePermissions,
) Keeper {
	k := Keeper{
		cdc:         cdc,
		storeKey:    storeKey,
		ccsStorage:  ccsKeeper,
		modulePerms: types.NewModulePerms(),
	}
	for _, requester := range permsRequesters {
		k.modulePerms.AutoAddRequester(requester)
	}

	return k
}
