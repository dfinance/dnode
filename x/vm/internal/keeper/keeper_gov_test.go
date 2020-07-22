// +build unit

package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dfinance/dnode/x/vm/internal/types"
)

type proposalInput struct {
	id       uint64
	proposal types.TestProposal
}

func TestVMKeeper_GovProposalQueue(t *testing.T) {
	input := newTestInput(true)
	defer input.Stop()

	testProposals := []proposalInput{
		{0, types.NewTestProposal(100, 150)},
		{1, types.NewTestProposal(200, 250)},
		{2, types.NewTestProposal(300, 350)},
	}

	cmpProposals := func(testProposalsIdx, rcvId uint64, rcvProposal types.PlannedProposal) {
		rcvP, ok := rcvProposal.(types.TestProposal)
		require.True(t, ok, "idx [%d]: type assert", testProposalsIdx)

		testProposal := testProposals[testProposalsIdx]

		require.Equal(t, testProposal.id, rcvId, "idx [%d]: indices", testProposalsIdx)
		require.Equal(t, testProposal.proposal.Value, rcvP.Value, "idx [%d]: value", testProposalsIdx)
		require.Equal(t, testProposal.proposal.ProposalType(), rcvP.ProposalType(), "idx [%d]: type", testProposalsIdx)
		require.Equal(t, testProposal.proposal.GetPlan(), rcvProposal.GetPlan(), "idx [%d]: plan", testProposalsIdx)
	}

	// add proposals to queue
	for _, p := range testProposals {
		require.NotEqual(t, input.vk.ScheduleProposal(input.ctx, p.proposal), "adding proposal %d", p.id)
	}

	// check all proposals exist
	{
		i := uint64(0)
		input.vk.IterateProposalsQueue(input.ctx, func(id uint64, p types.PlannedProposal) {
			cmpProposals(i, id, p)
			i++
		})
	}

	// check removing proposal
	{
		rmIdx := uint64(1)
		testProposals = append(testProposals[0:rmIdx], testProposals[rmIdx+1:]...)
		input.vk.RemoveProposalFromQueue(input.ctx, rmIdx)

		i := uint64(0)
		input.vk.IterateProposalsQueue(input.ctx, func(id uint64, p types.PlannedProposal) {
			cmpProposals(i, id, p)
			i++
		})
	}

	// check removing all
	{
		testProposals = nil
		input.vk.RemoveProposalFromQueue(input.ctx, 0)
		input.vk.RemoveProposalFromQueue(input.ctx, 2)
		input.vk.RemoveProposalFromQueue(input.ctx, 3) // removing non-existing

		cnt := 0
		input.vk.IterateProposalsQueue(input.ctx, func(_ uint64, _ types.PlannedProposal) {
			cnt++
		})
		require.Zero(t, cnt)
	}

	// check adding one
	{
		newProposal := types.NewTestProposal(400, 450)
		testProposals = append(testProposals, proposalInput{3, newProposal})
		require.NoError(t, input.vk.ScheduleProposal(input.ctx, newProposal), "adding new proposal")

		i := uint64(0)
		input.vk.IterateProposalsQueue(input.ctx, func(id uint64, p types.PlannedProposal) {
			cmpProposals(i, id, p)
			i++
		})
		require.Equal(t, uint64(1), i)

	}
}
