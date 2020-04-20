package types

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/x/gov"
)

type PlannedProposal struct {
	Proposal gov.Content  `json:"proposal"`
	Data     ProposalData `json:"data"`
	Plan     Plan         `json:"plan"`
}

func (p PlannedProposal) String() string {
	return fmt.Sprintf(`PlannedProposal:
  Proposal: %s
  Data: %s
  Plan: %s
`, p.Proposal.String(), p.Data.String(), p.Plan.String())
}

func (p PlannedProposal) ValidateBasic() error {
	if err := p.Proposal.ValidateBasic(); err != nil {
		return fmt.Errorf("proposal: %w", err)
	}

	if err := p.Plan.ValidateBasic(); err != nil {
		return fmt.Errorf("plan: %w", err)
	}

	return nil
}

type ProposalData interface {
	fmt.Stringer
	IsProposalData()
}

func NewPlannedProposal(proposal gov.Content, data ProposalData, plan Plan) PlannedProposal {
	return PlannedProposal{
		Proposal: proposal,
		Data:     data,
		Plan:     plan,
	}
}
