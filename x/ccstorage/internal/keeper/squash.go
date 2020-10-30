package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type (
	// Operations order:
	//   1: supplyOp
	SquashOptions struct {
		// Supply modification operations
		supplyOps supplyOperation
	}

	supplyOperation struct {
		// Set supply amount to zero
		SetToZero bool
	}
)

func (opts *SquashOptions) SetSupplyOperation(toZero bool) error {
	op := supplyOperation{
		SetToZero: toZero,
	}
	opts.supplyOps = op

	return nil
}

func NewEmptySquashOptions() SquashOptions {
	return SquashOptions{
		supplyOps: supplyOperation{},
	}
}

// PrepareForZeroHeight squashes current context state to fit zero-height (used on genesis export).
func (k Keeper) PrepareForZeroHeight(ctx sdk.Context, opts SquashOptions) error {
	// supplyOps
	{
		if opts.supplyOps.SetToZero {
			for _, cur := range k.GetCurrencies(ctx) {
				cur.Supply = sdk.ZeroInt()
				k.storeCurrency(ctx, cur)
				k.storeResStdCurrencyInfo(ctx, cur)
			}
		}
	}

	return nil
}
