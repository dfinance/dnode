package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/dfinance/dnode/x/vm/internal/types"
)

// ScheduleProposal checks if proposal can be added to gov proposal queue and adds it.
func (keeper Keeper) ScheduleProposal(ctx sdk.Context, pProposal types.PlannedProposal) error {
	keeper.modulePerms.AutoCheck(types.PermInit)

	if pProposal.GetPlan().Height <= ctx.BlockHeight() {
		return sdkErrors.Wrapf(
			sdkErrors.ErrInvalidRequest,
			"proposal can't be scheduled, planned blockHeight LE than current: %d le %d",
			pProposal.GetPlan().Height, ctx.BlockHeight())
	}

	keeper.addProposalToQueue(ctx, pProposal)

	return nil
}

// RemoveProposalFromQueue removes proposal from the gov proposal queue.
func (keeper Keeper) RemoveProposalFromQueue(ctx sdk.Context, id uint64) {
	keeper.modulePerms.AutoCheck(types.PermInit)

	queueKey := types.GetProposalQueueKey(id)

	store := ctx.KVStore(keeper.storeKey)
	store.Delete(queueKey)
}

// IterateProposalsQueue iterates over gov proposal queue.
func (keeper Keeper) IterateProposalsQueue(ctx sdk.Context, handler func(id uint64, pProposal types.PlannedProposal)) {
	keeper.modulePerms.AutoCheck(types.PermInit)

	iterator := keeper.proposalQueueIterator(ctx)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		id := types.SplitProposalQueueKey(iterator.Key())

		p, err := keeper.unmarshalPlannedProposal(iterator.Value())
		if err != nil {
			panic(err)
		}

		handler(id, p)
	}
}

// getNextProposalID returns next gov proposal queue ID.
func (keeper Keeper) getNextProposalID(ctx sdk.Context) (id uint64) {
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

// setProposalID updates gov proposal queue last ID.
func (keeper Keeper) setProposalID(ctx sdk.Context, id uint64) {
	store := ctx.KVStore(keeper.storeKey)

	bz := keeper.cdc.MustMarshalBinaryLengthPrefixed(id)
	store.Set(types.ProposalIDKey, bz)
}

// addProposalToQueue adds proposal to the gov proposal queue.
func (keeper Keeper) addProposalToQueue(ctx sdk.Context, pProposal types.PlannedProposal) {
	id := keeper.getNextProposalID(ctx)
	queueKey := types.GetProposalQueueKey(id)

	store := ctx.KVStore(keeper.storeKey)
	bz := keeper.cdc.MustMarshalBinaryLengthPrefixed(pProposal)
	store.Set(queueKey, bz)
	keeper.setProposalID(ctx, id)
}

// proposalQueueIterator returns gov proposal queue iterator.
func (keeper Keeper) proposalQueueIterator(ctx sdk.Context) sdk.Iterator {
	store := ctx.KVStore(keeper.storeKey)
	return store.Iterator(types.ProposalQueuePrefix, sdk.PrefixEndBytes(types.ProposalQueuePrefix))
}

// unmarshalPlannedProposal unmarshals stored PlannedProposal to a known concrete type.
func (keeper Keeper) unmarshalPlannedProposal(bz []byte) (types.PlannedProposal, error) {
	pStdlib := types.StdlibUpdateProposal{}
	if err := keeper.cdc.UnmarshalBinaryLengthPrefixed(bz, &pStdlib); err == nil {
		return pStdlib, nil
	}

	pTest := types.TestProposal{}
	if err := keeper.cdc.UnmarshalBinaryLengthPrefixed(bz, &pTest); err == nil {
		return pTest, nil
	}

	return nil, sdkErrors.Wrap(types.ErrInternal, "unkown stored PlannedProposal type")
}
