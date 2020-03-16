package poa

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	msTypes "github.com/dfinance/dnode/x/multisig/types"
	"github.com/dfinance/dnode/x/poa/types"
)

// New message handler for PoA module.
func NewHandler(keeper Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		return msTypes.ErrOnlyMultisig(types.DefaultCodespace, types.ModuleName).Result()
	}
}
