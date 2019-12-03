package currencies

import (
	cdcCodec "github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"wings-blockchain/x/currencies/types"
)

// Currency keeper struct
type Keeper struct {
	coinKeeper bank.Keeper
	cdc        *cdcCodec.Codec
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
func (keeper Keeper) DestroyCurrency(ctx sdk.Context, chainID, symbol, recipient string, amount sdk.Int, spender sdk.AccAddress) sdk.Error {
	if !keeper.doesCurrencyExists(ctx, symbol) {
		return sdk.ErrInsufficientCoins("no known coins to destroy")
	}

	keeper.reduceSupply(ctx, chainID, symbol, recipient, amount, spender)

	newCoin := sdk.NewCoin(symbol, amount)

	_, err := keeper.coinKeeper.SubtractCoins(ctx, spender, sdk.Coins{newCoin})

	return err
}

// Issue currency
func (keeper Keeper) IssueCurrency(ctx sdk.Context, symbol string, amount sdk.Int, decimals int8, recipient sdk.AccAddress, issueID string) sdk.Error {
	if keeper.hasIssue(ctx, issueID) {
		return types.ErrExistsIssue(issueID)
	}

	var isNew bool

	if isNew = keeper.doesCurrencyExists(ctx, symbol); !isNew {
		currency := types.NewCurrency(symbol, amount, decimals)
		keeper.storeCurrency(ctx, currency)
	} else {
		currency := keeper.getCurrency(ctx, symbol)

		if currency.Decimals != decimals {
			return types.ErrIncorrectDecimals(currency.Decimals, decimals, symbol)
		}

		keeper.increaseSupply(ctx, symbol, amount)
	}

	issue := types.NewIssue(symbol, amount, recipient)

	keeper.storeIssue(ctx, issueID, issue)

	newCoin := sdk.NewCoin(symbol, amount)

	_, err := keeper.coinKeeper.AddCoins(ctx, recipient, sdk.Coins{newCoin})

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

// Get codec
func (keeper Keeper) GetCDC() *cdcCodec.Codec {
	return keeper.cdc
}

// Get currency issue by id
func (keeper Keeper) GetIssue(ctx sdk.Context, issueID string) types.Issue {
	store := ctx.KVStore(keeper.storeKey)

	bz := store.Get(types.GetIssuesKey(issueID))

	var issue types.Issue
	keeper.cdc.MustUnmarshalBinaryBare(bz, &issue)

	return issue
}

// Has destroy
func (keeper Keeper) HasDestroy(ctx sdk.Context, id sdk.Int) bool {
	store := ctx.KVStore(keeper.storeKey)

	return store.Has(types.GetDestroyKey(id))
}

// Get destroy by id
func (keeper Keeper) GetDestroy(ctx sdk.Context, id sdk.Int) types.Destroy {
	store := ctx.KVStore(keeper.storeKey)

	var destroy types.Destroy
	keeper.cdc.MustUnmarshalBinaryBare(store.Get(types.GetDestroyKey(id)), &destroy)

	return destroy
}

// Checking does currency exists by symbol
func (keeper Keeper) doesCurrencyExists(ctx sdk.Context, symbol string) bool {
	store := ctx.KVStore(keeper.storeKey)
	return store.Has(types.GetCurrencyKey(symbol))
}

// Increase currency supply by symbol
func (keeper Keeper) increaseSupply(ctx sdk.Context, symbol string, amount sdk.Int) {
	currency := keeper.getCurrency(ctx, symbol)
	currency.Supply = currency.Supply.Add(amount)

	keeper.storeCurrency(ctx, currency)
}

// Reduce currency supply by symbol
func (keeper Keeper) reduceSupply(ctx sdk.Context, chainID, symbol, recipient string, amount sdk.Int, spender sdk.AccAddress) {
	currency := keeper.getCurrency(ctx, symbol)
	currency.Supply = currency.Supply.Sub(amount)

	newId := keeper.getNewID(ctx)
	destroy := types.NewDestroy(newId, chainID, symbol, amount, spender, recipient, ctx.TxBytes(), ctx.BlockHeader().Time.Unix())

	keeper.storeDestroy(ctx, destroy)
	keeper.storeCurrency(ctx, currency)
	keeper.setLastID(ctx, newId)
}

// Store destroy
func (keeper Keeper) storeDestroy(ctx sdk.Context, destroy types.Destroy) {
	store := ctx.KVStore(keeper.storeKey)
	store.Set(types.GetDestroyKey(destroy.ID), keeper.cdc.MustMarshalBinaryBare(destroy))
}

// Set last ID
func (keeper Keeper) setLastID(ctx sdk.Context, lastId sdk.Int) {
	store := ctx.KVStore(keeper.storeKey)
	store.Set(types.GetLastIDKey(), keeper.cdc.MustMarshalBinaryBare(lastId))
}

// Get last id
func (keeper Keeper) getLastID(ctx sdk.Context) sdk.Int {
	store := ctx.KVStore(keeper.storeKey)
	var lastId sdk.Int
	keeper.cdc.MustUnmarshalBinaryBare(store.Get(types.GetLastIDKey()), &lastId)
	return lastId
}

// Get new id
func (keeper Keeper) getNewID(ctx sdk.Context) sdk.Int {
	store := ctx.KVStore(keeper.storeKey)

	if !store.Has(types.GetLastIDKey()) {
		return sdk.NewInt(0)
	}

	lastId := keeper.getLastID(ctx)
	return lastId.AddRaw(1)
}

// Store currency in storage
func (keeper Keeper) storeCurrency(ctx sdk.Context, currency types.Currency) {
	store := ctx.KVStore(keeper.storeKey)
	store.Set(types.GetCurrencyKey(currency.Symbol), keeper.cdc.MustMarshalBinaryBare(currency))
}

// Get currency from storage
func (keeper Keeper) getCurrency(ctx sdk.Context, symbol string) types.Currency {
	store := ctx.KVStore(keeper.storeKey)

	bz := store.Get(types.GetCurrencyKey(symbol))

	var currency types.Currency
	keeper.cdc.MustUnmarshalBinaryBare(bz, &currency)

	return currency
}

// Store currency issue by id
func (keeper Keeper) storeIssue(ctx sdk.Context, issueID string, issue types.Issue) {
	store := ctx.KVStore(keeper.storeKey)
	store.Set(types.GetIssuesKey(issueID), keeper.cdc.MustMarshalBinaryBare(issue))
}

// Check if issue exists by id
func (keeper Keeper) hasIssue(ctx sdk.Context, issueID string) bool {
	store := ctx.KVStore(keeper.storeKey)
	return store.Has(types.GetIssuesKey(issueID))
}
