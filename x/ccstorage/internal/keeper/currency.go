package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/dfinance/dnode/x/ccstorage/internal/types"
	"github.com/dfinance/dnode/x/common_vm"
)

// CreateCurrency creates a new currency object with VM resources.
func (k Keeper) CreateCurrency(ctx sdk.Context, params types.CurrencyParams) error {
	k.modulePerms.AutoCheck(types.PermCreate)

	denom := params.Denom
	if k.HasCurrency(ctx, denom) {
		return sdkErrors.Wrapf(types.ErrWrongDenom, "currency %q: exists", denom)
	}

	// build currency objects
	currency := types.NewCurrency(params, sdk.ZeroInt())
	_, err := types.NewResCurrencyInfo(currency, common_vm.StdLibAddress)
	if err != nil {
		return sdkErrors.Wrapf(types.ErrWrongParams, "currency %q: %v", denom, err)
	}

	// store VM path objects
	k.storeCurrencyBalancePath(ctx, denom, currency.BalancePath())
	k.storeCurrencyInfoPath(ctx, denom, currency.InfoPath())

	// store currency objects
	k.storeCurrency(ctx, currency)
	k.storeResStdCurrencyInfo(ctx, currency)

	ctx.EventManager().EmitEvent(types.NewCCCreatedEvent(currency))

	return nil
}

// HasCurrency checks that currency exists.
func (k Keeper) HasCurrency(ctx sdk.Context, denom string) bool {
	k.modulePerms.AutoCheck(types.PermRead)

	store := ctx.KVStore(k.storeKey)

	return store.Has(types.GetCurrencyKey(denom))
}

// GetCurrencies returns all registered currencies.
func (k Keeper) GetCurrencies(ctx sdk.Context) types.Currencies {
	k.modulePerms.AutoCheck(types.PermRead)

	currencies := types.Currencies{}
	store := ctx.KVStore(k.storeKey)

	iterator := sdk.KVStorePrefixIterator(store, types.GetCurrencyKeyPrefix())
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var currency types.Currency
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &currency)

		currencies = append(currencies, currency)
	}

	return currencies
}

// GetCurrency returns currency.
func (k Keeper) GetCurrency(ctx sdk.Context, denom string) (types.Currency, error) {
	k.modulePerms.AutoCheck(types.PermRead)

	if !k.HasCurrency(ctx, denom) {
		return types.Currency{}, sdkErrors.Wrapf(types.ErrWrongDenom, "currency with %q denom: not found", denom)
	}

	return k.getCurrency(ctx, denom), nil
}

// IncreaseCurrencySupply increases currency supply and updates VM resources.
func (k Keeper) IncreaseCurrencySupply(ctx sdk.Context, coin sdk.Coin) error {
	k.modulePerms.AutoCheck(types.PermUpdate)

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
	k.modulePerms.AutoCheck(types.PermUpdate)

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
