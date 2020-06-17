package types

import (
	"fmt"
	"net/url"
	"strings"

	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/gov"
)

const (
	ProposalTypeStdlibUpdate = "StdlibUpdate"
)

var (
	_ gov.Content     = StdlibUpdateProposal{}
	_ PlannedProposal = StdlibUpdateProposal{}
)

// StdlibUpdateProposal is a gov proposal used to update DVM stdlib code.
type StdlibUpdateProposal struct {
	// Stdlib update source URL
	Url string `json:"url"`
	// Update description
	UpdateDescription string `json:"update_description"`
	// Proposal plan
	Plan Plan `json:"plan"`
	// Stdlib update bytecode
	Code []byte `json:"code"`
}

func (p StdlibUpdateProposal) GetTitle() string       { return "DVM Stdlib update" }
func (p StdlibUpdateProposal) GetDescription() string { return "Updates DVM stdlib code" }
func (p StdlibUpdateProposal) ProposalRoute() string  { return GovRouterKey }
func (p StdlibUpdateProposal) ProposalType() string   { return ProposalTypeStdlibUpdate }
func (p StdlibUpdateProposal) GetPlan() Plan          { return p.Plan }

func (p StdlibUpdateProposal) ValidateBasic() error {
	if err := p.Plan.ValidateBasic(); err != nil {
		return sdkErrors.Wrapf(ErrGovInvalidProposal, "plan: %v", err)
	}

	if p.Url == "" {
		return sdkErrors.Wrapf(ErrGovInvalidProposal, "url: empty")
	}
	if _, err := url.Parse(p.Url); err != nil {
		return sdkErrors.Wrapf(ErrGovInvalidProposal, "url: %v", err)
	}
	if p.UpdateDescription == "" {
		return sdkErrors.Wrapf(ErrGovInvalidProposal, "updateDescription: empty")
	}
	if len(p.Code) == 0 {
		return sdkErrors.Wrapf(ErrGovInvalidProposal, "code: empty")
	}

	return nil
}

func (p StdlibUpdateProposal) String() string {
	b := strings.Builder{}
	b.WriteString("Proposal:\n")
	b.WriteString(fmt.Sprintf("  Title: %s\n", p.GetTitle()))
	b.WriteString(fmt.Sprintf("  Description: %s\n", p.GetDescription()))
	b.WriteString(fmt.Sprintf("  %s", p.Plan.String()))
	b.WriteString(fmt.Sprintf("  Source URL: %s\n", p.Url))
	b.WriteString(fmt.Sprintf("  Update description: %s\n", p.UpdateDescription))

	return b.String()
}

// NewStdlibUpdateProposal creates a StdlibUpdateProposal object.
func NewStdlibUpdateProposal(plan Plan, url, updateDescription string, byteCode []byte) gov.Content {
	return StdlibUpdateProposal{
		Plan:              plan,
		Url:               url,
		UpdateDescription: updateDescription,
		Code:              byteCode,
	}
}
