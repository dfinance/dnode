package poa

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/dfinance/dnode/x/core"
	"github.com/dfinance/dnode/x/poa/internal/keeper"
)

// NewHandler creates sdk.Msg type messages handler.
func NewHandler(k keeper.Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		return nil, sdkErrors.Wrap(core.ErrNotMultisigModule, ModuleName)
	}
}
