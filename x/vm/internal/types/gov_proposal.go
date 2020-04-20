package types

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/x/gov"
)

const (
	ProposalTypeModuleUpdate = "ModuleUpdate"
	ProposalTypeTest         = "Test"
)

type ModuleUpdateProposal struct {
	Plan Plan            `json:"plan"`
	Msg  MsgDeployModule `json:"msg"`
}

func (p ModuleUpdateProposal) GetTitle() string       { return "Module update" }
func (p ModuleUpdateProposal) GetDescription() string { return "Updates stdlib module" }
func (p ModuleUpdateProposal) ProposalRoute() string  { return GovRouterKey }
func (p ModuleUpdateProposal) ProposalType() string   { return ProposalTypeModuleUpdate }
func (p ModuleUpdateProposal) GetPlan() Plan          { return p.Plan }
func (p ModuleUpdateProposal) ValidateBasic() error {
	if err := p.Plan.ValidateBasic(); err != nil {
		return fmt.Errorf("plan: %w", err)
	}
	if err := p.Msg.ValidateBasic(); err != nil {
		return fmt.Errorf("msg: %w", err)
	}

	return nil
}

func (p ModuleUpdateProposal) String() string {
	return fmt.Sprintf(`Proposal:
  Title: %s
  Description: %s
  Plan: %s
`, p.GetTitle(), p.GetDescription(), p.Plan.String())
}

func NewModuleUpdateProposal(plan Plan) gov.Content {
	return ModuleUpdateProposal{
		Plan: plan,
	}
}

type TestProposal struct {
	Plan  Plan   `json:"plan"`
	Value string `json:"value"`
}

func (p TestProposal) GetTitle() string       { return "Test" }
func (p TestProposal) GetDescription() string { return "Test proposal" }
func (p TestProposal) ProposalRoute() string  { return GovRouterKey }
func (p TestProposal) ProposalType() string   { return ProposalTypeTest }
func (p TestProposal) GetPlan() Plan          { return p.Plan }
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
  Plan: %s
`, p.GetTitle(), p.GetDescription(), p.Value, p.Plan.String())
}

func NewTestProposal(plan Plan, value string) gov.Content {
	return TestProposal{
		Plan:  plan,
		Value: value,
	}
}

func init() {
	gov.RegisterProposalType(ProposalTypeModuleUpdate)
	gov.RegisterProposalTypeCodec(ModuleUpdateProposal{}, ModuleName+"/ModuleUpdateProposal")
	gov.RegisterProposalType(ProposalTypeTest)
	gov.RegisterProposalTypeCodec(TestProposal{}, ModuleName+"/TestProposal")
}
