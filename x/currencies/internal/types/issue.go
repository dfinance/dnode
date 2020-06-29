package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Issue is an info about issuing currency to the payee (recipient).
type Issue struct {
	// Target currency denom
	Denom  string         `json:"denom" example:"dfi"`
	// Amount of coins payee balance is increased to
	Amount sdk.Int        `json:"amount" swaggertype:"string" example:"100"`
	// Target account for increasing coin balance
	Payee  sdk.AccAddress `json:"payee" swaggertype:"string" format:"bech32" example:"wallet13jyjuz3kkdvqw8u4qfkwd94emdl3vx394kn07h"`
}

func (issue Issue) String() string {
	return fmt.Sprintf("Issue:\n"+
		"  Denom:  %s\n"+
		"  Amount: %s\n"+
		"  Payee:  %s",
		issue.Denom,
		issue.Amount.String(),
		issue.Payee.String(),
	)
}

// NewIssue creates a new Issue object.
func NewIssue(denom string, amount sdk.Int, payee sdk.AccAddress) Issue {
	return Issue{
		Denom:  denom,
		Amount: amount,
		Payee:  payee,
	}
}
