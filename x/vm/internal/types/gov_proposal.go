package types

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/x/gov"
)

const (
	ProposalTypeTest = "Test"
)

type TestProposal struct {
	Value string
}

func (c TestProposal) GetTitle() string       { return "Test content title" }
func (c TestProposal) GetDescription() string { return "Test content description" }
func (c TestProposal) ProposalRoute() string  { return GovRouterKey }
func (c TestProposal) ProposalType() string   { return ProposalTypeTest }
func (c TestProposal) ValidateBasic() error   { return nil }
func (c TestProposal) String() string {
	return fmt.Sprintf("%s: %s (%s): %s", c.ProposalType(), c.GetTitle(), c.GetDescription(), c.Value)
}

func NewTestProposal(value string) gov.Content {
	return TestProposal{Value: value}
}

func init() {
	gov.RegisterProposalType(ProposalTypeTest)
	gov.RegisterProposalTypeCodec(TestProposal{}, ModuleName+"/TestProposal")
}