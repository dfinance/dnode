package currencies

import (
	"github.com/cosmos/cosmos-sdk/x/bank"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cdcCodec "github.com/cosmos/cosmos-sdk/codec"
)

// Currency keeper struct
type Keeper struct {
	coinKeeper bank.Keeper
	cdc 	   *cdcCodec.Codec
	storeKey   sdk.StoreKey
}

// Create new currency keeper
func NewKeeper(coinKeeper bank.Keeper, storeKey sdk.StoreKey, cdc *cdcCodec.Codec) Keeper {
	return Keeper{
		coinKeeper: coinKeeper,
		storeKey:   storeKey,
		cdc:        cdc,
	}
}

// Destroy currency
func (keeper Keeper) DestroyCurrency(ctx sdk.Context, symbol string, amount int64, sender sdk.AccAddress) sdk.Error {
	store := ctx.KVStore(keeper.storeKey)

	if !keeper.doesCurrencyExists(store, symbol) {
		return sdk.ErrInsufficientCoins("no known coins to destroy")
	}

	keeper.reduceSupply(store, symbol, amount)

	newCoin := sdk.NewInt64Coin(symbol, amount)

	_, _, err := keeper.coinKeeper.SubtractCoins(ctx, sender, sdk.Coins{newCoin})

	return err
}

// Issue currency
func (keeper Keeper) IssueCurrency(ctx sdk.Context, symbol string, amount int64, decimals int8, creator sdk.AccAddress) sdk.Error {
	store := ctx.KVStore(keeper.storeKey)

	var isNew bool

	if isNew = keeper.doesCurrencyExists(store, symbol); !isNew {
		currency := NewCurrency(symbol, amount, decimals, creator)
		keeper.storeCurrency(store, currency)
	} else {
		keeper.increaseSupply(store, symbol, amount)
	}

	newCoin := sdk.NewInt64Coin(symbol, amount)

	_, _, err := keeper.coinKeeper.AddCoins(ctx, creator, sdk.Coins{newCoin})

	return err
}

// Checking does currency exists by symbol
func (keeper Keeper) doesCurrencyExists(store sdk.KVStore, symbol string) bool {
	return store.Has([]byte(symbol))
}

// Increase currency supply by symbol
func (keeper Keeper) increaseSupply(store sdk.KVStore, symbol string, amount int64) {
	currency := keeper.getCurrency(store, symbol)

	currency.Supply += amount

	keeper.storeCurrency(store, currency)
}

// Reduce currency supply by symbol
func (keeper Keeper) reduceSupply(store sdk.KVStore, symbol string, amount int64) {
	currency := keeper.getCurrency(store, symbol)

	currency.Supply -= amount

	keeper.storeCurrency(store, currency)
}

// Store currency in storage
func (keeper Keeper) storeCurrency(store sdk.KVStore, currency Currency) {
	store.Set([]byte(currency.Symbol), keeper.cdc.MustMarshalBinaryBare(currency))
}

// Get currency from storage
func (keeper Keeper) getCurrency(store sdk.KVStore, symbol string) Currency {
	bz := store.Get([]byte(symbol))

	var currency Currency
	keeper.cdc.MustUnmarshalBinaryBare(bz, &currency)

	return currency
}

// Remove currency
func (keeper Keeper) removeCurrency(store sdk.KVStore, symbol string) {
	store.Delete([]byte(symbol))
}
