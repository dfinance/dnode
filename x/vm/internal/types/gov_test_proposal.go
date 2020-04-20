package types

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/x/gov"
)

type TestProposal struct {
	Plan  Plan   `json:"plan"`
	Value string `json:"value"`
}

func (p TestProposal) GetTitle() string       { return "Test" }
func (p TestProposal) GetDescription() string { return "Test proposal" }
func (p TestProposal) ProposalRoute() string  { return GovRouterKey }
func (p TestProposal) ProposalType() string   { return ProposalTypeTest }
func (p TestProposal) ValidateBasic() error {
	if err := p.Plan.ValidateBasic(); err != nil {
		return fmt.Errorf("plan: %w", err)
	}
	if p.Value == "" {
		return fmt.Errorf("value: empty")
	}

	return nil
}

func (p TestProposal) String() string {
	return fmt.Sprintf(`Proposal:
  Title: %s
  Description: %s
  Value: %s
`, p.GetTitle(), p.GetDescription(), p.Value)
}

func NewTestProposal(plan Plan, value string) gov.Content {
	return TestProposal{
		Plan:  plan,
		Value: value,
	}
}

func init() {
	gov.RegisterProposalType(ProposalTypeTest)
	gov.RegisterProposalTypeCodec(TestProposal{}, ModuleName+"/TestProposal")
}
