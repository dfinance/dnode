// Implements custom AnteHandler.
package core

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/supply"
	"wings-blockchain/cmd/config"
)

var (
	DefaultFees = sdk.Coins{sdk.NewCoin(config.MainDenom, sdk.NewInt(1))}
)

// Custom antehandler catches and prevents transactions without fees and fees not in "wings" currency
// After execution of custom logic, call standard auth.AnteHandler.
func NewAnteHandler(ak auth.AccountKeeper, supplyKeeper supply.Keeper, sigGasConsumer auth.SignatureVerificationGasConsumer) sdk.AnteHandler {
	return func(ctx sdk.Context, tx sdk.Tx, simulate bool) (newCtx sdk.Context, res sdk.Result, abort bool) {
		stdTx, ok := tx.(auth.StdTx)

		if !ok {
			newCtx = auth.SetGasMeter(simulate, ctx, 0)
			return newCtx, sdk.ErrInternal("tx must be StdTx").Result(), true
		}

		// ignore genesis block.
		if ctx.BlockHeight() > 0 {
			if stdTx.Fee.Amount.IsZero() {
				newCtx = auth.SetGasMeter(simulate, ctx, 0)
				return newCtx, sdk.ErrInternal("tx must contains fees").Result(), true
			}

			if ctx.BlockHeight() != 0 && !stdTx.Fee.Amount.DenomsSubsetOf(DefaultFees) {
				newCtx = auth.SetGasMeter(simulate, ctx, 0)
				return newCtx, sdk.ErrInternal(fmt.Sprintf("tx must contains fees only in %s denom", config.MainDenom)).Result(), true
			}
		}

		return auth.NewAnteHandler(ak, supplyKeeper, sigGasConsumer)(ctx, tx, simulate)
	}
}
