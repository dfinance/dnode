package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/dfinance/dnode/x/vm/internal/types"
)

// ScheduleProposal checks if proposal can be added to gov proposal queue and adds it.
func (k Keeper) ScheduleProposal(ctx sdk.Context, pProposal types.PlannedProposal) error {
	k.modulePerms.AutoCheck(types.PermInit)

	if pProposal.GetPlan().Height <= ctx.BlockHeight() {
		return sdkErrors.Wrapf(
			sdkErrors.ErrInvalidRequest,
			"proposal can't be scheduled, planned blockHeight LE than current: %d le %d",
			pProposal.GetPlan().Height, ctx.BlockHeight())
	}

	k.addProposalToQueue(ctx, pProposal)

	return nil
}

// RemoveProposalFromQueue removes proposal from the gov proposal queue.
func (k Keeper) RemoveProposalFromQueue(ctx sdk.Context, id uint64) {
	k.modulePerms.AutoCheck(types.PermInit)

	queueKey := types.GetProposalQueueKey(id)

	store := ctx.KVStore(k.storeKey)
	store.Delete(queueKey)
}

// IterateProposalsQueue iterates over gov proposal queue.
func (k Keeper) IterateProposalsQueue(ctx sdk.Context, handler func(id uint64, pProposal types.PlannedProposal)) {
	k.modulePerms.AutoCheck(types.PermInit)

	iterator := k.proposalQueueIterator(ctx)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		id := types.SplitProposalQueueKey(iterator.Key())

		p, err := k.unmarshalPlannedProposal(iterator.Value())
		if err != nil {
			panic(err)
		}

		handler(id, p)
	}
}

// getNextProposalID returns next gov proposal queue ID.
func (k Keeper) getNextProposalID(ctx sdk.Context) (id uint64) {
	store := ctx.KVStore(k.storeKey)
	if !store.Has(types.ProposalIDKey) {
		return 0
	}

	bz := store.Get(types.ProposalIDKey)
	if err := k.cdc.UnmarshalBinaryLengthPrefixed(bz, &id); err != nil {
		panic(err)
	}

	return id + 1
}

// setProposalID updates gov proposal queue last ID.
func (k Keeper) setProposalID(ctx sdk.Context, id uint64) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshalBinaryLengthPrefixed(id)
	store.Set(types.ProposalIDKey, bz)
}

// addProposalToQueue adds proposal to the gov proposal queue.
func (k Keeper) addProposalToQueue(ctx sdk.Context, pProposal types.PlannedProposal) {
	id := k.getNextProposalID(ctx)
	queueKey := types.GetProposalQueueKey(id)

	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(pProposal)
	store.Set(queueKey, bz)
	k.setProposalID(ctx, id)
}

// proposalQueueIterator returns gov proposal queue iterator.
func (k Keeper) proposalQueueIterator(ctx sdk.Context) sdk.Iterator {
	store := ctx.KVStore(k.storeKey)
	return store.Iterator(types.ProposalQueuePrefix, sdk.PrefixEndBytes(types.ProposalQueuePrefix))
}

// unmarshalPlannedProposal unmarshals stored PlannedProposal to a known concrete type.
func (k Keeper) unmarshalPlannedProposal(bz []byte) (types.PlannedProposal, error) {
	pStdlib := types.StdlibUpdateProposal{}
	if err := k.cdc.UnmarshalBinaryLengthPrefixed(bz, &pStdlib); err == nil {
		return pStdlib, nil
	}

	pTest := types.TestProposal{}
	if err := k.cdc.UnmarshalBinaryLengthPrefixed(bz, &pTest); err == nil {
		return pTest, nil
	}

	return nil, sdkErrors.Wrap(types.ErrInternal, "unkown stored PlannedProposal type")
}
