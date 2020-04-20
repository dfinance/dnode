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
`, p.GetTitle(), p.GetDescription())
}

func NewModuleUpdateProposal(plan Plan) gov.Content {
	return ModuleUpdateProposal{
		Plan: plan,
	}
}

func init() {
	gov.RegisterProposalType(ProposalTypeModuleUpdate)
	gov.RegisterProposalTypeCodec(ModuleUpdateProposal{}, ModuleName+"/ModuleUpdateProposal")
}
