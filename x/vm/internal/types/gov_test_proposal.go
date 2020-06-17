package types

import "github.com/cosmos/cosmos-sdk/x/gov"

var (
	_ gov.Content     = TestProposal{}
	_ PlannedProposal = TestProposal{}
)

// TestProposal is only used for unit tests.
type TestProposal struct {
	Plan  Plan
	Value int
}

func (p TestProposal) GetTitle() string       { return "Test title" }
func (p TestProposal) GetDescription() string { return "Test description" }
func (p TestProposal) ProposalRoute() string  { return "Test_route" }
func (p TestProposal) ProposalType() string   { return "Test_proposal_type" }
func (p TestProposal) GetPlan() Plan          { return p.Plan }
func (p TestProposal) ValidateBasic() error   { return nil }
func (p TestProposal) String() string         { return "" }

func NewTestProposal(value int, blockHeight int64) TestProposal {
	return TestProposal{
		Value: value,
		Plan:  NewPlan(blockHeight),
	}
}
