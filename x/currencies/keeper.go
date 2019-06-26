package currencies

import (
	"github.com/cosmos/cosmos-sdk/x/bank"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cdcCodec "github.com/cosmos/cosmos-sdk/codec"
	"wings-blockchain/x/currencies/types"
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
func (keeper Keeper) DestroyCurrency(ctx sdk.Context, symbol string, amount sdk.Int, sender sdk.AccAddress) sdk.Error {
	store := ctx.KVStore(keeper.storeKey)

	if !keeper.doesCurrencyExists(store, symbol) {
		return sdk.ErrInsufficientCoins("no known coins to destroy")
	}

	keeper.reduceSupply(store, symbol, amount)

	newCoin := sdk.NewCoin(symbol, amount)

	_, _, err := keeper.coinKeeper.SubtractCoins(ctx, sender, sdk.Coins{newCoin})

	return err
}

// Issue currency
func (keeper Keeper) IssueCurrency(ctx sdk.Context, symbol string, amount sdk.Int, decimals int8, recipient sdk.AccAddress, issueID string) sdk.Error {
	store := ctx.KVStore(keeper.storeKey)

	if keeper.hasIssue(store, issueID) {
	    return types.ErrExistsIssue(issueID)
    }

	var isNew bool

	if isNew = keeper.doesCurrencyExists(store, symbol); !isNew {
		currency := types.NewCurrency(symbol, amount, decimals)
		keeper.storeCurrency(store, currency)
	} else {
	    currency := keeper.getCurrency(store, symbol)

	    if currency.Decimals != decimals {
	        return types.ErrIncorrectDecimals(currency.Decimals, decimals, symbol)
        }

		keeper.increaseSupply(store, symbol, amount)
	}

	issue := types.NewIssue(symbol, amount, recipient)

	keeper.storeIssue(store, issueID, issue)

	newCoin := sdk.NewCoin(symbol, amount)

	_, _, err := keeper.coinKeeper.AddCoins(ctx, recipient, sdk.Coins{newCoin})

	return err
}


// Get currency by denom/symbol
func (keeper Keeper) GetCurrency(ctx sdk.Context, symbol string) types.Currency {
	store := ctx.KVStore(keeper.storeKey)

	var currency types.Currency
	bs := store.Get(types.GetCurrencyKey(symbol))

	keeper.cdc.MustUnmarshalBinaryBare(bs, &currency)

	return currency
}

// Checking does currency exists by symbol
func (keeper Keeper) doesCurrencyExists(store sdk.KVStore, symbol string) bool {
	return store.Has([]byte(symbol))
}

// Increase currency supply by symbol
func (keeper Keeper) increaseSupply(store sdk.KVStore, symbol string, amount sdk.Int) {
	currency := keeper.getCurrency(store, symbol)

	currency.Supply = currency.Supply.Add(amount)

	keeper.storeCurrency(store, currency)
}

// Reduce currency supply by symbol
func (keeper Keeper) reduceSupply(store sdk.KVStore, symbol string, amount sdk.Int) {
	currency := keeper.getCurrency(store, symbol)

	currency.Supply = currency.Supply.Sub(amount)

	keeper.storeCurrency(store, currency)
}

// Store currency in storage
func (keeper Keeper) storeCurrency(store sdk.KVStore, currency types.Currency) {
	store.Set(types.GetCurrencyKey(currency.Symbol), keeper.cdc.MustMarshalBinaryBare(currency))
}

// Get currency from storage
func (keeper Keeper) getCurrency(store sdk.KVStore, symbol string) types.Currency {
	bz := store.Get(types.GetCurrencyKey(symbol))

	var currency types.Currency
	keeper.cdc.MustUnmarshalBinaryBare(bz, &currency)

	return currency
}

// Get codec
func (keeper Keeper) GetCDC() *cdcCodec.Codec {
	return keeper.cdc
}

// Get currency issue by id
func (keeper Keeper) GetIssue(ctx sdk.Context, issueID string) types.Issue {
    store := ctx.KVStore(keeper.storeKey)

    bz    := store.Get(types.GetIssuesKey(issueID))

    var issue types.Issue
    keeper.cdc.MustUnmarshalBinaryBare(bz, &issue)

    return issue
}

// Store currency issue by id
func (keeper Keeper) storeIssue(store sdk.KVStore, issueID string, issue types.Issue) {
    store.Set(types.GetIssuesKey(issueID), keeper.cdc.MustMarshalBinaryBare(issue))
}

// Check if issue exists by id
func (keeper Keeper) hasIssue(store sdk.KVStore, issueID string) bool {
    return store.Has(types.GetIssuesKey(issueID))
}
