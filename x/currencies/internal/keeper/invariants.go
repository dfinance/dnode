package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dfinance/dnode/x/currencies/internal/types"
)

// RegisterInvariants registers all module invariants
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	ir.RegisterRoute(types.ModuleName, "total-supply", TotalSupply(k))
}

// TotalSupply checks that the currency total supply and supply module amounts are equal.
func TotalSupply(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		supplyCoins := sdk.NewCoins()
		if supply := k.supplyKeeper.GetSupply(ctx); supply != nil {
			supplyCoins = supplyCoins.Add(supply.GetTotal()...)
		}
		supplyCoins.Sort()

		currenciesCoins := sdk.NewCoins()
		for denom := range k.ccsKeeper.GetCurrenciesParams(ctx) {
			currency, err := k.ccsKeeper.GetCurrency(ctx, denom)
			if err != nil {
				panic(fmt.Errorf("currency %q read failed", denom))
			}

			currenciesCoins = currenciesCoins.Add(currency.GetSupplyCoin())
		}
		currenciesCoins.Sort()

		broken := !currenciesCoins.IsEqual(currenciesCoins)
		irComment := fmt.Sprintf(
			"\tccStorage.Supplies: %s\n\tsupply.Supplies: %s\n",
			currenciesCoins.String(), currenciesCoins.String(),
		)

		return sdk.FormatInvariant(types.ModuleName, "total-supply", irComment), broken
	}
}
