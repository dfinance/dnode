package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/multisig/internal/types"
)

// ConfirmCall confirms (adds vote) call by {address}.
func (k Keeper) ConfirmCall(ctx sdk.Context, id dnTypes.ID, address sdk.AccAddress) error {
	call, err := k.GetCall(ctx, id)
	if err != nil {
		return err
	}

	return k.storeVote(ctx, call, address)
}

// RevokeConfirmation revokes {address} confirmation (removes vote) for the call.
func (k Keeper) RevokeConfirmation(ctx sdk.Context, id dnTypes.ID, address sdk.AccAddress) error {
	call, err := k.GetCall(ctx, id)
	if err != nil {
		return err
	}

	return k.revokeVote(ctx, call, address)
}

// HasVote checks that call is confirmed by {address}.
func (k Keeper) HasVote(ctx sdk.Context, callID dnTypes.ID, address sdk.AccAddress) (bool, error) {
	store := ctx.KVStore(k.storeKey)

	if !store.Has(types.GetVotesKey(callID)) {
		return false, sdkErrors.Wrap(types.ErrVoteNoVotes, callID.String())
	}

	var votes types.Votes
	bs := store.Get(types.GetVotesKey(callID))
	k.cdc.MustUnmarshalBinaryBare(bs, &votes)

	for _, vote := range votes {
		if vote.Equals(address) {
			return true, nil
		}
	}

	return false, nil
}

// GetVotes returns vote for specific call.
func (k Keeper) GetVotes(ctx sdk.Context, id dnTypes.ID) (types.Votes, error) {
	store := ctx.KVStore(k.storeKey)

	if !k.HasCall(ctx, id) {
		return types.Votes{}, sdkErrors.Wrap(types.ErrWrongCallId, id.String())
	}

	votesKey := types.GetVotesKey(id)
	if !store.Has(votesKey) {
		return types.Votes{}, nil
	}

	var votes types.Votes
	bz := store.Get(votesKey)
	k.cdc.MustUnmarshalBinaryBare(bz, &votes)

	return votes, nil
}

// GetConfirmationsCount returns number of confirmations for specific call.
func (k Keeper) GetConfirmationsCount(ctx sdk.Context, id dnTypes.ID) (uint64, error) {
	votes, err := k.GetVotes(ctx, id)

	return uint64(len(votes)), err
}

// storeVote checks that {address} can vote for {call} and adds it to the storage.
func (k Keeper) storeVote(ctx sdk.Context, call types.Call, address sdk.AccAddress) (retErr error) {
	defer func() {
		if retErr == nil {
			ctx.EventManager().EmitEvent(types.NewConfirmVoteEvent(call.ID, address))
		}
	}()

	store := ctx.KVStore(k.storeKey)

	if err := call.CanBeVoted(); err != nil {
		return err
	}

	// check if address has already voted
	voteExists, err := k.HasVote(ctx, call.ID, address)
	if voteExists {
		return sdkErrors.Wrapf(types.ErrVoteAlreadyConfirmed, "%s by %s", call.ID.String(), address.String())
	}

	// check if call has no votes yet
	votesKey := types.GetVotesKey(call.ID)
	if types.ErrVoteNoVotes.Is(err) {
		votes := types.Votes{address}
		store.Set(votesKey, k.cdc.MustMarshalBinaryBare(votes))

		return nil
	}

	// append vote to existing votes
	var votes types.Votes
	bz := store.Get(votesKey)
	k.cdc.MustUnmarshalBinaryBare(bz, &votes)

	votes = append(votes, address)
	store.Set(votesKey, k.cdc.MustMarshalBinaryBare(votes))

	return nil
}

// revokeVote check that {address} can revoke his vote from the call and removes it from the storage.
func (k Keeper) revokeVote(ctx sdk.Context, call types.Call, address sdk.AccAddress) (retErr error) {
	defer func() {
		if retErr == nil {
			ctx.EventManager().EmitEvent(types.NewRevokeVoteEvent(call.ID, address))
		}
	}()

	store := ctx.KVStore(k.storeKey)

	if err := call.CanBeVoted(); err != nil {
		return err
	}

	// check if address has approved this call
	voteExists, err := k.HasVote(ctx, call.ID, address)
	if err != nil || !voteExists {
		return sdkErrors.Wrapf(types.ErrVoteNotApproved, "%s by %s", call.ID.String(), address.String())
	}

	votesKey := types.GetVotesKey(call.ID)
	var votes types.Votes
	bz := store.Get(votesKey)
	k.cdc.MustUnmarshalBinaryBare(bz, &votes)

	// remove votes if this is the last vote
	if len(votes) == 1 {
		store.Delete(votesKey)
		k.RemoveCallFromQueue(ctx, call.ID, call.Height)

		return nil
	}

	// remove vote from existing votes
	voteIdx := -1
	for i, vote := range votes {
		if vote.Equals(address) {
			voteIdx = i
			break
		}
	}

	votes = append(votes[:voteIdx], votes[voteIdx+1:]...)
	store.Set(votesKey, k.cdc.MustMarshalBinaryBare(votes))

	return nil
}
