package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dfinance/dnode/x/cc_storage"
)

// CreateCurrency redirects CreateCurrency request to the currencies storage.
func (k Keeper) CreateCurrency(ctx sdk.Context, denom string, params cc_storage.CurrencyParams) error {
	return k.ccsKeeper.CreateCurrency(ctx, denom, params)
}
