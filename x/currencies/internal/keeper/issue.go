package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/dfinance/dnode/x/currencies/internal/types"
)

// IssueCurrency issues a new currency and increases payee coin balance.
// Issue is a multisig operation.
func (k Keeper) IssueCurrency(ctx sdk.Context, id string, coin sdk.Coin, payee sdk.AccAddress) (retErr error) {
	k.modulePerms.AutoCheck(types.PermIssue)

	// bankKeeper might panic
	defer func() {
		if r := recover(); r != nil {
			retErr = sdkErrors.Wrapf(types.ErrInternal, "bankKeeper.AddCoins for address %q panic: %v", payee.String(), r)
		}
	}()

	if k.stakingKeeper.IsAccountBanned(ctx, payee) {
		return sdkErrors.Wrapf(types.ErrAccountBanned, "account banned: %s", payee.String())
	}

	if k.HasIssue(ctx, id) {
		return sdkErrors.Wrapf(types.ErrWrongIssueID, "issue with ID %q: already exists", id)
	}

	// check and update currency
	if _, err := k.ccsKeeper.GetCurrency(ctx, coin.Denom); err != nil {
		return err
	}

	// store issue
	issue := types.NewIssue(coin, payee)
	k.storeIssue(ctx, id, issue)

	// update account balance
	if _, err := k.bankKeeper.AddCoins(ctx, payee, sdk.Coins{coin}); err != nil {
		return sdkErrors.Wrapf(types.ErrInternal, "bankKeeper.AddCoins for address %q: %v", payee.String(), err)
	}

	// increase supply
	if err := k.ccsKeeper.IncreaseCurrencySupply(ctx, coin); err != nil {
		return err
	}

	curSupply := k.supplyKeeper.GetSupply(ctx)
	curSupply = curSupply.SetTotal(curSupply.GetTotal().Add(coin))
	k.supplyKeeper.SetSupply(ctx, curSupply)

	ctx.EventManager().EmitEvent(types.NewIssueEvent(id, coin, payee))

	return
}

// HasIssue checks that issue exists.
func (k Keeper) HasIssue(ctx sdk.Context, id string) bool {
	k.modulePerms.AutoCheck(types.PermRead)

	store := ctx.KVStore(k.storeKey)

	return store.Has(types.GetIssuesKey(id))
}

// GetIssue returns issue.
func (k Keeper) GetIssue(ctx sdk.Context, id string) (types.Issue, error) {
	k.modulePerms.AutoCheck(types.PermRead)

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

// getGenesisIssues returns all registered issues with meta (GenesisIssue) from the storage.
func (k Keeper) getGenesisIssues(ctx sdk.Context) []types.GenesisIssue {
	issues := make([]types.GenesisIssue, 0)

	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.GetIssuesPrefix())
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var issue types.Issue
		k.cdc.MustUnmarshalBinaryBare(iterator.Value(), &issue)

		issueID := types.MustParseIssueKey(iterator.Key())
		issues = append(issues, types.GenesisIssue{
			Issue: issue,
			ID:    issueID,
		})
	}

	return issues
}

// storeIssue sets issue to the storage.
func (k Keeper) storeIssue(ctx sdk.Context, id string, issue types.Issue) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetIssuesKey(id), k.cdc.MustMarshalBinaryBare(issue))
}
