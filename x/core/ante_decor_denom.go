package core

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth"

	"github.com/dfinance/dnode/cmd/config/genesis/defaults"
)

// DenomDecorator catches and prevents transactions without fees and fees not in "xfi" currency
type DenomDecorator struct{}

func NewDenomDecorator() DenomDecorator {
	return DenomDecorator{}
}

func (dd DenomDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	stdTx, ok := tx.(auth.StdTx)
	if !ok {
		newCtx = auth.SetGasMeter(simulate, ctx, 0)
		err = sdkErrors.Wrap(ErrInternal, "tx must be StdTx")
		return
	}

	// ignore genesis block.
	if ctx.BlockHeight() > 0 {
		if stdTx.Fee.Amount.IsZero() {
			return auth.SetGasMeter(simulate, ctx, 0), ErrFeeRequired
		}

		if !stdTx.Fee.Amount.DenomsSubsetOf(DefaultFees) {
			return auth.SetGasMeter(simulate, ctx, 0), sdkErrors.Wrap(ErrWrongFeeDenom, defaults.MainDenom)
		}
	}

	return next(ctx, tx, simulate)
}
