package poa

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	msTypes "wings-blockchain/x/multisig/types"
	"wings-blockchain/x/poa/types"
)

// New message handler for PoA module.
func NewHandler(keeper Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		return msTypes.ErrOnlyMultisig(types.DefaultCodespace, types.ModuleName).Result()
	}
}
