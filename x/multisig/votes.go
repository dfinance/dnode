// Keeper votes part (votes managment).
package multisig

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/dfinance/dnode/x/multisig/types"
)

// Confirm call.
func (keeper Keeper) Confirm(ctx sdk.Context, id uint64, address sdk.AccAddress) error {
	return keeper.storeVote(ctx, id, address)
}

// Revoke confirmation from call.
func (keeper Keeper) RevokeConfirmation(ctx sdk.Context, id uint64, address sdk.AccAddress) error {
	return keeper.revokeVote(ctx, id, address)
}

// Get votes for specific call.
func (keeper Keeper) GetVotes(ctx sdk.Context, id uint64) (types.Votes, error) {
	store := ctx.KVStore(keeper.storeKey)

	if !store.Has(types.GetKeyVotesById(id)) {
		return types.Votes{}, sdkErrors.Wrapf(types.ErrWrongCallId, "%d", id)
	}

	var votes types.Votes
	bs := store.Get(types.GetKeyVotesById(id))

	keeper.cdc.MustUnmarshalBinaryBare(bs, &votes)

	return votes, nil
}

// Get message confirmations.
func (keeper Keeper) GetConfirmations(ctx sdk.Context, id uint64) (uint64, error) {
	votes, err := keeper.GetVotes(ctx, id)

	return uint64(len(votes)), err
}

// Check if message confirmed by address.
func (keeper Keeper) HasVote(ctx sdk.Context, id uint64, address sdk.AccAddress) (bool, error) {
	store := ctx.KVStore(keeper.storeKey)

	if !store.Has(types.GetKeyVotesById(id)) {
		return false, sdkErrors.Wrapf(types.ErrNoVotes, "%d", id)
	}

	var votes types.Votes
	bs := store.Get(types.GetKeyVotesById(id))

	keeper.cdc.MustUnmarshalBinaryBare(bs, &votes)

	for _, vote := range votes {
		if vote.Equals(address) {
			return true, nil
		}
	}

	return false, nil
}

// Store vote for message by address.
func (keeper Keeper) storeVote(ctx sdk.Context, id uint64, address sdk.AccAddress) error {
	store := ctx.KVStore(keeper.storeKey)

	nextId := keeper.getNextCallId(ctx)

	if id > nextId-1 {
		return sdkErrors.Wrapf(types.ErrWrongCallId, "%d", id)
	}

	call := keeper.getCallById(ctx, id)

	if call.Approved {
		return sdkErrors.Wrapf(types.ErrAlreadyConfirmed, "%d", id)
	}

	if call.Rejected {
		return sdkErrors.Wrapf(types.ErrAlreadyRejected, "%d", id)
	}

	if has, _ := keeper.HasVote(ctx, id, address); has {
		return sdkErrors.Wrapf(types.ErrCallAlreadyApproved, "%d by %s", id, address.String())
	}

	if !store.Has(types.GetKeyVotesById(id)) {
		votes := types.Votes{address}
		store.Set(types.GetKeyVotesById(id), keeper.cdc.MustMarshalBinaryBare(votes))

		return nil
	}

	var votes types.Votes
	bs := store.Get(types.GetKeyVotesById(id))

	keeper.cdc.MustUnmarshalBinaryBare(bs, &votes)
	votes = append(votes, address)

	store.Set(types.GetKeyVotesById(id), keeper.cdc.MustMarshalBinaryBare(votes))

	return nil
}

// Revoke confirmation from message by address.
func (keeper Keeper) revokeVote(ctx sdk.Context, id uint64, address sdk.AccAddress) error {
	store := ctx.KVStore(keeper.storeKey)

	if !store.Has(types.GetKeyVotesById(id)) {
		return sdkErrors.Wrapf(types.ErrNoVotes, "%d", id)
	}

	call := keeper.getCallById(ctx, id)

	if call.Approved {
		return sdkErrors.Wrapf(types.ErrAlreadyConfirmed, "%d", id)
	}

	if call.Rejected {
		return sdkErrors.Wrapf(types.ErrAlreadyRejected, "%d", id)
	}

	if has, _ := keeper.HasVote(ctx, id, address); !has {
		return sdkErrors.Wrapf(types.ErrCallNotApproved, "%d by %s", id, address.String())
	}

	var votes types.Votes
	bs := store.Get(types.GetKeyVotesById(id))

	keeper.cdc.MustUnmarshalBinaryBare(bs, &votes)

	if len(votes) == 1 {
		store.Delete(types.GetKeyVotesById(id))
		keeper.removeCallFromQueue(ctx, id, call.Height)

		return nil
	}

	index := -1
	for i, vote := range votes {
		if vote.Equals(address) {
			index = i
			break
		}
	}

	votes = append(votes[:index], votes[index+1:]...)
	store.Set(types.GetKeyVotesById(id), keeper.cdc.MustMarshalBinaryBare(votes))

	return nil
}
