// +build unit

package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dfinance/dnode/x/vm/internal/types"
)

func Test_GovProposalQueue(t *testing.T) {
	input := newTestInput(true)
	defer input.Stop()

	testProposals := []types.PlannedProposal{
		types.NewPlainProposal(100, 150),
		types.NewPlainProposal(200, 250),
		types.NewPlainProposal(300, 350),
	}

	cmpProposals := func(testProposalsIdx, rcvId uint64, rcvProposal types.PlannedProposal) {
		rcvP, ok := rcvProposal.(types.TestProposal)
		require.True(t, ok, "idx [%d]: type assert", testProposalsIdx)

		testProposal := testProposals[testProposalsIdx]
		testP := testProposal.(types.TestProposal)

		require.Equal(t, testProposalsIdx, rcvId, "idx [%d]: indices", testProposalsIdx)
		require.Equal(t, testP.Value, rcvP.Value, "idx [%d]: value", testProposalsIdx)
		require.Equal(t, testP.ProposalType(), rcvP.ProposalType(), "idx [%d]: type", testProposalsIdx)
		require.Equal(t, testProposal.GetPlan(), rcvProposal.GetPlan(), "idx [%d]: plan", testProposalsIdx)
	}

	// add proposals to queue
	for _, p := range testProposals {
		input.vk.ScheduleProposal(input.ctx, p)
	}

	// check all proposals exist
	{
		i := uint64(0)
		input.vk.IterateProposalsQueue(input.ctx, func(id uint64, p types.PlannedProposal) bool {
			cmpProposals(i, id, p)
			i++
			return false
		})
	}

	// check removing proposal
	{
		rmIdx := uint64(1)
		testProposals = append(testProposals[0:rmIdx], testProposals[rmIdx+1:]...)
		input.vk.RemoveProposalFromQueue(input.ctx, rmIdx)

		i := uint64(0)
		input.vk.IterateProposalsQueue(input.ctx, func(id uint64, p types.PlannedProposal) bool {
			cmpProposals(i, id, p)
			i++
			return false
		})
	}

	// check removing all
	{
		input.vk.RemoveProposalFromQueue(input.ctx, 0)
		input.vk.RemoveProposalFromQueue(input.ctx, 2)
		input.vk.RemoveProposalFromQueue(input.ctx, 3)

		cnt := 0
		input.vk.IterateProposalsQueue(input.ctx, func(_ uint64, _ types.PlannedProposal) bool {
			cnt++
			return false
		})
		require.Zero(t, cnt)
	}

	// check adding one
	{
		input.vk.ScheduleProposal(input.ctx, types.NewPlainProposal(400, 450))

		cnt := 0
		input.vk.IterateProposalsQueue(input.ctx, func(_ uint64, _ types.PlannedProposal) bool {
			cnt++
			return false
		})
		require.Equal(t, 1, cnt)
	}
}
