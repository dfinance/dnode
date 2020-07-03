package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/dfinance/dnode/x/currencies/internal/types"
)

// IssueCurrency issues a new currency and increases payee coin balance.
// Issue is a multisig operation.
func (k Keeper) IssueCurrency(ctx sdk.Context, id, denom string, amount sdk.Int, decimals uint8, payee sdk.AccAddress) (retErr error) {
	// bankKeeper might panic
	defer func() {
		if r := recover(); r != nil {
			retErr = sdkErrors.Wrapf(types.ErrInternal, "bankKeeper.AddCoins for address %q panic: %v", payee, r)
		}
	}()

	if k.HasIssue(ctx, id) {
		return sdkErrors.Wrapf(types.ErrWrongIssueID, "issue with ID %q: already exists", id)
	}

	// check and update currency
	currency, err := k.ccsKeeper.GetCurrency(ctx, denom)
	if err != nil {
		return err
	}
	if currency.Decimals != decimals {
		return sdkErrors.Wrapf(types.ErrIncorrectDecimals, "currency %q decimals: %d", denom, currency.Decimals)
	}

	if err := k.ccsKeeper.IncreaseCurrencySupply(ctx, denom, amount); err != nil {
		return err
	}

	// store issue
	issue := types.NewIssue(denom, amount, payee)
	k.storeIssue(ctx, id, issue)

	// update account balance
	newCoin := sdk.NewCoin(denom, amount)
	if _, err := k.bankKeeper.AddCoins(ctx, payee, sdk.Coins{newCoin}); err != nil {
		return sdkErrors.Wrapf(types.ErrInternal, "bankKeeper.AddCoins for address %q: %v", payee, err)
	}

	return
}

// HasIssue checks that issue exists.
func (k Keeper) HasIssue(ctx sdk.Context, id string) bool {
	store := ctx.KVStore(k.storeKey)

	return store.Has(types.GetIssuesKey(id))
}

// GetIssue returns issue.
func (k Keeper) GetIssue(ctx sdk.Context, id string) (types.Issue, error) {
	if !k.HasIssue(ctx, id) {
		return types.Issue{}, sdkErrors.Wrapf(types.ErrWrongIssueID, "issueID %q: not found", id)
	}

	return k.getIssue(ctx, id), nil
}

// getIssue returns issue from the storage.
func (k Keeper) getIssue(ctx sdk.Context, id string) types.Issue {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetIssuesKey(id))

	issue := types.Issue{}
	k.cdc.MustUnmarshalBinaryBare(bz, &issue)

	return issue
}

// storeIssue sets issue to the storage.
func (k Keeper) storeIssue(ctx sdk.Context, id string, issue types.Issue) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetIssuesKey(id), k.cdc.MustMarshalBinaryBare(issue))
}
