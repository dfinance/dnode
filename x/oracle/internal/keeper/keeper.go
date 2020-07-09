// Oracle module keeper creates and stores assets, prices and oracle objects.
package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"

	"github.com/dfinance/dnode/x/common_vm"
	"github.com/dfinance/dnode/x/oracle/internal/types"
)

// Keeper struct for oracle module
type Keeper struct {
	storeKey   sdk.StoreKey        // The keys used to access the stores from Context
	cdc        *codec.Codec        // Codec for binary encoding/decoding
	paramstore params.Subspace     // The reference to the Paramstore to get and set oracle specific params
	vmKeeper   common_vm.VMStorage // Virtual machine keeper
}

// IsNominee checks is nominee exist in the keeper params.
func (k Keeper) IsNominee(ctx sdk.Context, nominee string) bool {
	p := k.GetParams(ctx)
	nominees := p.Nominees
	for _, v := range nominees {
		if v == nominee {
			return true
		}
	}
	return false
}

// NewKeeper returns a new keeper for the oralce module.
func NewKeeper(
	storeKey sdk.StoreKey,
	cdc *codec.Codec,
	paramstore params.Subspace,
	vmKeeper common_vm.VMStorage,
) Keeper {
	return Keeper{
		paramstore: paramstore.WithKeyTable(types.ParamKeyTable()),
		storeKey:   storeKey,
		cdc:        cdc,
		vmKeeper:   vmKeeper,
	}
}
