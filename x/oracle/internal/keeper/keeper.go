// Oracle module keeper creates and stores assets, prices and oracle objects.
package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"

	"github.com/dfinance/dnode/helpers/perms"
	"github.com/dfinance/dnode/x/common_vm"
	"github.com/dfinance/dnode/x/oracle/internal/types"
)

// Keeper struct for oracle module
type Keeper struct {
	storeKey    sdk.StoreKey        // The keys used to access the stores from Context
	cdc         *codec.Codec        // Codec for binary encoding/decoding
	paramstore  params.Subspace     // The reference to the Paramstore to get and set oracle specific params
	vmKeeper    common_vm.VMStorage // Virtual machine keeper
	modulePerms perms.ModulePermissions
}

// IsNominee checks is nominee exist in the keeper params.
func (k Keeper) IsNominee(ctx sdk.Context, address string) error {
	k.modulePerms.AutoCheck(types.PermReader)

	p := k.GetParams(ctx)
	for _, v := range p.Nominees {
		if v == address {
			return nil
		}
	}

	return fmt.Errorf("address %q is not a nominee", address)
}

// NewKeeper returns a new keeper for the oralce module.
func NewKeeper(
	cdc *codec.Codec,
	storeKey sdk.StoreKey,
	paramStore params.Subspace,
	vmKeeper common_vm.VMStorage,
	permsRequesters ...perms.RequestModulePermissions,
) Keeper {
	k := Keeper{
		cdc:         cdc,
		storeKey:    storeKey,
		paramstore:  paramStore.WithKeyTable(types.ParamKeyTable()),
		vmKeeper:    vmKeeper,
		modulePerms: types.NewModulePerms(),
	}
	for _, requester := range permsRequesters {
		k.modulePerms.AutoAddRequester(requester)
	}

	return k
}
