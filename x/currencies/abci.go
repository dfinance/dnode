package currencies

import (
	"bytes"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dfinance/dnode/x/ccstorage"
	"github.com/dfinance/dnode/x/currencies/internal/types"
)

// Temporary solution to catch supply update and update it in ccstorage (so VM will have update total supply).
// TODO: solve supply keeper issue: create own supply keeper, that will also save supply to VM storage.
func BeginBlocker(ctx sdk.Context, ccsKeeper ccstorage.Keeper) {
	if ctx.BlockHeight() > 1 {
		for _, event := range ctx.EventManager().Events() {
			if event.Type == types.MintEventType {
				// found mint event.
				for _, attr := range event.Attributes {
					if bytes.Equal(attr.Key, []byte(sdk.AttributeKeyAmount)) {
						// found amount attribute, only amount passed, no denom, so use default one.
						amount, isOk := sdk.NewIntFromString(string(attr.Value))
						if !isOk {
							panic(fmt.Errorf("error during parsing mint event inflation amount: %q", string(attr.Value)))
						}

						coin := sdk.NewCoin(types.MintDenom, amount)
						if err := ccsKeeper.IncreaseCurrencySupply(ctx, coin); err != nil {
							panic(fmt.Errorf("error during increasing total supply by %q: %v", coin, err))
						}
						break
					}
				}
			}
		}
	}
}
