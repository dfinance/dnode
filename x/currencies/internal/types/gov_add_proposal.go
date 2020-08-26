package types

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/x/gov"

	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/ccstorage"
)

const (
	ProposalTypeAddCurrency = "AddCurrency"
)

var (
	_ gov.Content = AddCurrencyProposal{}
)

// AddCurrencyProposal is a gov proposal to add currency to the module.
type AddCurrencyProposal struct {
	Denom            string
	Decimals         uint8
	VmBalancePathHex string
	VmInfoPathHex    string
}

func (p AddCurrencyProposal) GetTitle() string       { return "Add currency" }
func (p AddCurrencyProposal) GetDescription() string { return "Creates new non-token currency" }
func (p AddCurrencyProposal) ProposalRoute() string  { return GovRouterKey }
func (p AddCurrencyProposal) ProposalType() string   { return ProposalTypeAddCurrency }

func (p AddCurrencyProposal) ValidateBasic() error {
	if err := dnTypes.DenomFilter(p.Denom); err != nil {
		return fmt.Errorf("denom: %w", err)
	}
	if err := p.GetCurrencyParams().Validate(); err != nil {
		return fmt.Errorf("params: %w", err)
	}

	return nil
}

func (p AddCurrencyProposal) GetCurrencyParams() ccstorage.CurrencyParams {
	return ccstorage.CurrencyParams{
		Denom:    p.Denom,
		Decimals: p.Decimals,
	}
}

func (p AddCurrencyProposal) String() string {
	b := strings.Builder{}
	b.WriteString("Proposal:\n")
	b.WriteString(fmt.Sprintf("  Title: %s\n", p.GetTitle()))
	b.WriteString(fmt.Sprintf("  Description: %s\n", p.GetDescription()))
	b.WriteString(fmt.Sprintf("  Denom: %s\n", p.Denom))
	b.WriteString(fmt.Sprintf("  Decimals: %d\n", p.Decimals))
	b.WriteString(fmt.Sprintf("  VmBalancePathHex: 0x%s\n", p.VmBalancePathHex))
	b.WriteString(fmt.Sprintf("  VmInfoPathHex: %s", p.VmInfoPathHex))

	return b.String()
}

// NewAddCurrencyProposal creates a AddCurrencyProposal object.
func NewAddCurrencyProposal(denom string, decimals uint8, balancePath, infoPath string) AddCurrencyProposal {
	return AddCurrencyProposal{
		Denom:            denom,
		Decimals:         decimals,
		VmBalancePathHex: balancePath,
		VmInfoPathHex:    infoPath,
	}
}
