package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/dfinance/dnode/x/vm/internal/types"
)

func (keeper Keeper) GetNextProposalID(ctx sdk.Context) (id uint64) {
	store := ctx.KVStore(keeper.storeKey)
	if !store.Has(types.ProposalIDKey) {
		return 0
	}

	bz := store.Get(types.ProposalIDKey)
	if err := keeper.cdc.UnmarshalBinaryLengthPrefixed(bz, &id); err != nil {
		panic(err)
	}

	return id + 1
}

func (keeper Keeper) SetProposalID(ctx sdk.Context, id uint64) {
	store := ctx.KVStore(keeper.storeKey)

	bz := keeper.cdc.MustMarshalBinaryLengthPrefixed(id)
	store.Set(types.ProposalIDKey, bz)
}

func (keeper Keeper) AddProposalToQueue(ctx sdk.Context, p types.PlannedProposal) {
	id := keeper.GetNextProposalID(ctx)
	queueKey := types.GetProposalQueueKey(id)

	store := ctx.KVStore(keeper.storeKey)
	bz := keeper.cdc.MustMarshalBinaryLengthPrefixed(p)
	store.Set(queueKey, bz)
	keeper.SetProposalID(ctx, id)
}

func (keeper Keeper) RemoveProposalFromQueue(ctx sdk.Context, id uint64) {
	queueKey := types.GetProposalQueueKey(id)

	store := ctx.KVStore(keeper.storeKey)
	store.Delete(queueKey)
}

func (keeper Keeper) ProposalQueueIterator(ctx sdk.Context) sdk.Iterator {
	store := ctx.KVStore(keeper.storeKey)
	return store.Iterator(types.ProposalQueuePrefix, sdk.PrefixEndBytes(types.ProposalQueuePrefix))
}

func (keeper Keeper) IterateProposalsQueue(ctx sdk.Context, handler func(id uint64, proposal types.PlannedProposal) (stop bool)) {
	iterator := keeper.ProposalQueueIterator(ctx)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		id := types.SplitProposalQueueKey(iterator.Key())

		p := types.PlannedProposal{}
		keeper.cdc.MustUnmarshalBinaryLengthPrefixed(iterator.Value(), &p)

		if !handler(id, p) {
			break
		}
	}
}

func (keeper Keeper) ScheduleProposal(ctx sdk.Context, p types.PlannedProposal) error {
	if p.Plan.Height <= ctx.BlockHeight() {
		return sdkErrors.Wrapf(
			sdkErrors.ErrInvalidRequest,
			"update can't be scheduled, planned blockHeight le than current: %d le %d",
			p.Plan.Height, ctx.BlockHeight())
	}

	keeper.AddProposalToQueue(ctx, p)

	return nil
}
