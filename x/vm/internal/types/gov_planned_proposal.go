package types

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/dfinance/dvm-proto/go/vm_grpc"
)

type PlannedProposal struct {
	Proposal gov.Content  `json:"proposal"`
	Data     ProposalData `json:"data"`
	Plan     Plan         `json:"plan"`
}

func (p PlannedProposal) String() string {
	return fmt.Sprintf(`PlannedProposal:
  %s
  %s
`, p.Proposal.String(), p.Plan.String())
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

type ProposalData struct {
	WriteSet []*vm_grpc.VMValue `json:"write_sets"`
	Events   []*vm_grpc.VMEvent `json:"events"`
}

func NewPlannedProposal(proposal gov.Content, data ProposalData, plan Plan) PlannedProposal {
	return PlannedProposal{
		Proposal: proposal,
		Data:     data,
		Plan:     plan,
	}
}
