// Implements custom AnteHandler.
package core

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	"wings-blockchain/cmd/config"
)

var (
	DefaultFees = sdk.Coins{sdk.NewCoin(config.MainDenom, sdk.NewInt(1))}
)

// Custom antehandler catches and prevents transactions without fees and fees not in "wings" currency
// After execution of custom logic, call standard auth.AnteHandler.
func NewAnteHandler(ak auth.AccountKeeper, supplyKeeper types.SupplyKeeper, sigGasConsumer auth.SignatureVerificationGasConsumer) sdk.AnteHandler {
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
				return newCtx, ErrFeeRequired().Result(), true
			}

			if !stdTx.Fee.Amount.DenomsSubsetOf(DefaultFees) {
				newCtx = auth.SetGasMeter(simulate, ctx, 0)
				return newCtx, ErrWrongFeeDenom(config.MainDenom).Result(), true
			}
		}

		return auth.NewAnteHandler(ak, supplyKeeper, sigGasConsumer)(ctx, tx, simulate)
	}
}
