package types

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/x/gov"
)

type PlannedProposal interface {
	gov.Content
	GetPlan() Plan
}

type ExecutableProposal struct {
	Type string          `json:"type"`
	Data PlannedProposal `json:"proposal"`
}

func (p ExecutableProposal) String() string {
	return fmt.Sprintf(`ExecutableProposal:
  Type: %s
  Data: %s
`, p.Type, p.Data.String())
}

func NewExecutableProposal(proposalType string, proposalData PlannedProposal) ExecutableProposal {
	return ExecutableProposal{
		Type: proposalType,
		Data: proposalData,
	}
}
