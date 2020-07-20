package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/dfinance/dnode/x/ccstorage/internal/types"
	"github.com/dfinance/dnode/x/common_vm"
)

// CreateCurrency creates a new currency object with VM resources.
func (k Keeper) CreateCurrency(ctx sdk.Context, denom string, params types.CurrencyParams) error {
	k.modulePerms.AutoCheck(types.PermCCCreator)

	if k.HasCurrency(ctx, denom) {
		return sdkErrors.Wrapf(types.ErrWrongDenom, "currency %q: exists", denom)
	}

	// build currency objects
	currency := types.NewCurrency(denom, sdk.ZeroInt(), params.Decimals)
	_, err := types.NewResCurrencyInfo(currency, common_vm.StdLibAddress)
	if err != nil {
		return sdkErrors.Wrapf(types.ErrWrongParams, "currency %q: %v", denom, err)
	}

	// store VM path objects
	k.storeCurrencyBalancePath(ctx, denom, params.BalancePath())
	k.storeCurrencyInfoPath(ctx, denom, params.InfoPath())

	// store currency objects
	k.storeCurrency(ctx, currency)
	k.storeResStdCurrencyInfo(ctx, currency)
	k.updateCurrenciesParams(ctx, denom, params)

	ctx.EventManager().EmitEvent(types.NewCCCreatedEvent(currency, params))

	return nil
}

// HasCurrency checks that currency exists.
func (k Keeper) HasCurrency(ctx sdk.Context, denom string) bool {
	k.modulePerms.AutoCheck(types.PermCCReader)

	store := ctx.KVStore(k.storeKey)

	return store.Has(types.GetCurrencyKey(denom))
}

// GetCurrency returns currency.
func (k Keeper) GetCurrency(ctx sdk.Context, denom string) (types.Currency, error) {
	k.modulePerms.AutoCheck(types.PermCCReader)

	if !k.HasCurrency(ctx, denom) {
		return types.Currency{}, sdkErrors.Wrapf(types.ErrWrongDenom, "currency with %q denom: not found", denom)
	}

	return k.getCurrency(ctx, denom), nil
}

// IncreaseCurrencySupply increases currency supply and updates VM resources.
func (k Keeper) IncreaseCurrencySupply(ctx sdk.Context, coin sdk.Coin) error {
	k.modulePerms.AutoCheck(types.PermCCUpdater)

	currency, err := k.GetCurrency(ctx, coin.Denom)
	if err != nil {
		return err
	}
	currency.Supply = currency.Supply.Add(coin.Amount)

	k.storeCurrency(ctx, currency)
	k.storeResStdCurrencyInfo(ctx, currency)

	return nil
}

// DecreaseCurrencySupply reduces currency supply and updates VM resources.
func (k Keeper) DecreaseCurrencySupply(ctx sdk.Context, coin sdk.Coin) error {
	k.modulePerms.AutoCheck(types.PermCCUpdater)

	currency, err := k.GetCurrency(ctx, coin.Denom)
	if err != nil {
		return err
	}
	currency.Supply = currency.Supply.Sub(coin.Amount)

	k.storeCurrency(ctx, currency)
	k.storeResStdCurrencyInfo(ctx, currency)

	return nil
}

// getCurrency returns currency from the storage
func (k Keeper) getCurrency(ctx sdk.Context, denom string) types.Currency {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetCurrencyKey(denom))

	currency := types.Currency{}
	k.cdc.MustUnmarshalBinaryBare(bz, &currency)

	return currency
}

// storeCurrency sets currency to the storage.
func (k Keeper) storeCurrency(ctx sdk.Context, currency types.Currency) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetCurrencyKey(currency.Denom), k.cdc.MustMarshalBinaryBare(currency))
}
