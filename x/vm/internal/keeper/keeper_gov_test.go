// +build unit

package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dfinance/dnode/x/vm/internal/types"
)

type PlainProposal struct {
	Value int
}

func (p PlainProposal) GetTitle() string       { return "Test title" }
func (p PlainProposal) GetDescription() string { return "Test description" }
func (p PlainProposal) ProposalRoute() string  { return "Test_route" }
func (p PlainProposal) ProposalType() string   { return "Test_proposal_type" }
func (p PlainProposal) ValidateBasic() error   { return nil }
func (p PlainProposal) String() string         { return "" }

func Test_GovProposalQueue(t *testing.T) {
	input := setupTestInput(true)
	defer closeInput(input)

	input.cdc.RegisterConcrete(PlainProposal{}, types.ModuleName+"/PlainProposal", nil)

	testProposals := []types.PlannedProposal{
		types.NewPlannedProposal(PlainProposal{Value: 100}, types.ProposalData{}, types.NewPlan(150)),
		types.NewPlannedProposal(PlainProposal{Value: 200}, types.ProposalData{}, types.NewPlan(250)),
		types.NewPlannedProposal(PlainProposal{Value: 300}, types.ProposalData{}, types.NewPlan(350)),
	}

	cmpProposals := func(testProposalsIdx, rcvId uint64, rcvProposal types.PlannedProposal) {
		rcvP, ok := rcvProposal.Proposal.(PlainProposal)
		require.True(t, ok, "idx [%d]: type assert", testProposalsIdx)

		testProposal := testProposals[testProposalsIdx]
		testP := testProposal.Proposal.(PlainProposal)

		require.Equal(t, testProposalsIdx, rcvId, "idx [%d]: indices", testProposalsIdx)
		require.Equal(t, testP.Value, rcvP.Value, "idx [%d]: value", testProposalsIdx)
		require.Equal(t, testP.ProposalType(), rcvP.ProposalType(), "idx [%d]: type", testProposalsIdx)
		require.Equal(t, testProposal.Plan, rcvProposal.Plan, "idx [%d]: plan", testProposalsIdx)
		require.Equal(t, testProposal.Data, rcvProposal.Data, "idx [%d]: data", testProposalsIdx)
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
		input.vk.ScheduleProposal(input.ctx, types.NewPlannedProposal(PlainProposal{Value: 400}, types.ProposalData{}, types.NewPlan(450)),)

		cnt := 0
		input.vk.IterateProposalsQueue(input.ctx, func(_ uint64, _ types.PlannedProposal) bool {
			cnt++
			return false
		})
		require.Equal(t, 1, cnt)
	}
}
