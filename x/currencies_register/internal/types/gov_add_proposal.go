package types

import (
	"encoding/hex"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
)

const (
	ProposalTypeAddCurrency = "AddCurrency"
)

var (
	_ gov.Content = AddCurrencyProposal{}
)

// AddCurrencyProposal is a gov proposal to add currency to the module.
type AddCurrencyProposal struct {
	Denom       string
	Decimals    uint8
	IsToken     bool
	Owner       sdk.AccAddress
	Path        []byte
	TotalSupply sdk.Int
}

func (p AddCurrencyProposal) GetTitle() string       { return "Add currency" }
func (p AddCurrencyProposal) GetDescription() string { return "Creates new currency" }
func (p AddCurrencyProposal) ProposalRoute() string  { return GovRouterKey }
func (p AddCurrencyProposal) ProposalType() string   { return ProposalTypeAddCurrency }

func (p AddCurrencyProposal) ValidateBasic() error {
	// Validation is skipped as it is done in NewCurrencyInfo func.
	return nil
}

func (p AddCurrencyProposal) String() string {
	b := strings.Builder{}
	b.WriteString("Proposal:\n")
	b.WriteString(fmt.Sprintf("  Title: %s\n", p.GetTitle()))
	b.WriteString(fmt.Sprintf("  Description: %s\n", p.GetDescription()))
	b.WriteString(fmt.Sprintf("  Denom: %s\n", p.Denom))
	b.WriteString(fmt.Sprintf("  Decimals: %d\n", p.Decimals))
	b.WriteString(fmt.Sprintf("  IsToken: %v\n", p.IsToken))
	b.WriteString(fmt.Sprintf("  Owner: %s\n", p.Owner.String()))
	b.WriteString(fmt.Sprintf("  Path: 0x%s\n", hex.EncodeToString(p.Path)))
	b.WriteString(fmt.Sprintf("  TotalSupply: %s\n", p.TotalSupply.String()))

	return b.String()
}

// NewAddCurrencyProposal creates a AddCurrencyProposal object.
func NewAddCurrencyProposal(denom string, decimals uint8, isToken bool, owner sdk.AccAddress, path []byte, totalSupply sdk.Int) AddCurrencyProposal {
	return AddCurrencyProposal{
		Denom:       denom,
		Decimals:    decimals,
		IsToken:     isToken,
		Owner:       owner,
		Path:        path,
		TotalSupply: totalSupply,
	}
}
