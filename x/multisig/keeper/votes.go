package keeper

import (
	"wings-blockchain/x/multisig/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Confirm call
func (keeper Keeper) Confirm(ctx sdk.Context, id uint64, address sdk.AccAddress) sdk.Error {
	err := keeper.storeVote(ctx, id, address)
	return err
}

// Revoke confirmation from call
func (keeper Keeper) RevokeConfirmation(ctx sdk.Context, id uint64, address sdk.AccAddress) sdk.Error {
	err := keeper.revokeVote(ctx, id, address)
	return err
}

// Get message confirmations
func (keeper Keeper) GetConfirmations(ctx sdk.Context, id uint64) (uint64, sdk.Error) {
	store := ctx.KVStore(keeper.storeKey)

	if !store.Has(types.GetKeyVotesById(id)) {
		return 0, types.ErrWrongCallId(id)
	}

	var votes types.Votes
	bs := store.Get(types.GetKeyVotesById(id))

	keeper.cdc.MustUnmarshalBinaryBare(bs, &votes)

	return uint64(len(votes)), nil
}

// Check if message confirmed by address
func (keeper Keeper) HasVote(ctx sdk.Context, id uint64, address sdk.AccAddress) (bool, sdk.Error) {
	store := ctx.KVStore(keeper.storeKey)

	if !store.Has(types.GetKeyVotesById(id)) {
		return false, types.ErrNoVotes(id)
	}

	var votes types.Votes
	bs := store.Get(types.GetKeyVotesById(id))

	keeper.cdc.MustUnmarshalBinaryBare(bs, &votes)

	for _, vote := range votes {
		if vote.Address.Equals(address) {
			return true, nil
		}
	}

	return false, nil
}

// Store vote for message by address
func (keeper Keeper) storeVote(ctx sdk.Context, id uint64, address sdk.AccAddress) sdk.Error {
	store := ctx.KVStore(keeper.storeKey)

	nextId := keeper.getNextCallId(ctx)

	if id > nextId-1 {
		return types.ErrWrongCallId(id)
	}

	call := keeper.getCallById(ctx, id)

	if call.Approved {
		return types.ErrAlreadyConfirmed(id)
	}

	if call.Rejected {
		return types.ErrAlreadyRejected(id)
	}

	if has, _ := keeper.HasVote(ctx, id, address); has {
		return types.ErrCallAlreadyApproved(id, address.String())
	}

	if !store.Has(types.GetKeyVotesById(id)) {
		votes := types.Votes{types.Vote{address}}
		store.Set(types.GetKeyVotesById(id), keeper.cdc.MustMarshalBinaryBare(votes))
		return nil
	}

	var votes types.Votes
	bs := store.Get(types.GetKeyVotesById(id))

	keeper.cdc.MustUnmarshalBinaryBare(bs, &votes)
	votes = append(votes, types.Vote{Address: address})

	store.Set(types.GetKeyVotesById(id), keeper.cdc.MustMarshalBinaryBare(votes))

	return nil
}

// Revoke confirmation from message by address
func (keeper Keeper) revokeVote(ctx sdk.Context, id uint64, address sdk.AccAddress) sdk.Error {
	store := ctx.KVStore(keeper.storeKey)

	if !store.Has(types.GetKeyVotesById(id)) {
		return types.ErrNoVotes(id)
	}

	call := keeper.getCallById(ctx, id)

	if call.Approved {
		return types.ErrAlreadyConfirmed(id)
	}

	if call.Rejected {
		return types.ErrAlreadyRejected(id)
	}

	if has, _ := keeper.HasVote(ctx, id, address); !has {
		return types.ErrCallNotApproved(id, address.String())
	}

	var votes types.Votes
	bs := store.Get(types.GetKeyVotesById(id))

	keeper.cdc.MustUnmarshalBinaryBare(bs, &votes)

	if len(votes) == 1 {
		votes = types.Votes{}
		store.Delete(types.GetKeyVotesById(id))
	} else {
		index := -1
		for i, vote := range votes {
			if vote.Address.Equals(address) {
				index = i
				break
			}
		}

		votes = append(votes[:index], votes[index+1:]...)
		store.Set(types.GetKeyVotesById(id), keeper.cdc.MustMarshalBinaryBare(votes))
	}


	return nil
}