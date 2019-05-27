package currencies

import (
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/types"
	cdcCodec "github.com/cosmos/cosmos-sdk/codec"
)

// Currency keeper struct
type Keeper struct {
	coinKeeper bank.Keeper
	cdc 	   *cdcCodec.Codec
	storeKey   types.StoreKey
}

// Create new currency keeper
func NewKeeper(coinKeeper bank.Keeper, storeKey types.StoreKey, cdc *cdcCodec.Codec) Keeper {
	return Keeper{
		coinKeeper: coinKeeper,
		storeKey:   storeKey,
		cdc:        cdc,
	}
}

// Destroy currency
func (keeper Keeper) DestroyCurrency(ctx types.Context, symbol string, amount int64, sender types.AccAddress) types.Error {
	store := ctx.KVStore(keeper.storeKey)

	if !keeper.doesCurrencyExists(store, symbol) {
		return types.ErrInsufficientCoins("no known coins to destroy")
	}

	keeper.reduceSupply(store, symbol, amount)

	newCoin := types.NewInt64Coin(symbol, amount)

	_, _, err := keeper.coinKeeper.SubtractCoins(ctx, sender, types.Coins{newCoin})

	if err != nil {
		keeper.increaseSupply(store, symbol, amount)
	}

	return err
}

// Issue currency
func (keeper Keeper) IssueCurrency(ctx types.Context, symbol string, amount int64, decimals int8, creator types.AccAddress) types.Error {
	store := ctx.KVStore(keeper.storeKey)

	var isNew bool

	if isNew = keeper.doesCurrencyExists(store, symbol); !isNew {
		currency := NewCurrency(symbol, amount, decimals, creator)
		keeper.storeCurrency(store, currency)
	} else {
		keeper.increaseSupply(store, symbol, amount)
	}

	newCoin := types.NewInt64Coin(symbol, amount)

	_, _, err := keeper.coinKeeper.AddCoins(ctx, creator, types.Coins{newCoin})

	if err != nil {
		if isNew {
			keeper.removeCurrency(store, symbol)
		} else {
			keeper.reduceSupply(store, symbol, amount)
		}
	}

	return err
}

// Checking does currency exists by symbol
func (keeper Keeper) doesCurrencyExists(store types.KVStore, symbol string) bool {
	return store.Has([]byte(symbol))
}

// Increase currency supply by symbol
func (keeper Keeper) increaseSupply(store types.KVStore, symbol string, amount int64) {
	currency := keeper.getCurrency(store, symbol)

	currency.Supply += amount

	keeper.storeCurrency(store, currency)
}

// Reduce currency supply by symbol
func (keeper Keeper) reduceSupply(store types.KVStore, symbol string, amount int64) {
	currency := keeper.getCurrency(store, symbol)

	currency.Supply -= amount

	keeper.storeCurrency(store, currency)
}

// Store currency in storage
func (keeper Keeper) storeCurrency(store types.KVStore, currency Currency) {
	store.Set([]byte(currency.Symbol), keeper.cdc.MustMarshalBinaryBare(currency))
}

// Get currency from storage
func (keeper Keeper) getCurrency(store types.KVStore, symbol string) Currency {
	bz := store.Get([]byte(symbol))

	var currency Currency
	keeper.cdc.MustUnmarshalBinaryBare(bz, &currency)

	return currency
}

// Remove currency
func (keeper Keeper) removeCurrency(store types.KVStore, symbol string) {
	store.Delete([]byte(symbol))
}
