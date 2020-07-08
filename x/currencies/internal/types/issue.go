package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Issue is an info about issuing currency to the payee (recipient).
type Issue struct {
	// Issuing coin
	Coin sdk.Coin `json:"coin" swaggertype:"string" example:"100dfi"`
	// Target account for increasing coin balance
	Payee sdk.AccAddress `json:"payee" swaggertype:"string" format:"bech32" example:"wallet13jyjuz3kkdvqw8u4qfkwd94emdl3vx394kn07h"`
}

func (issue Issue) String() string {
	return fmt.Sprintf("Issue:\n"+
		"  Coin:  %s\n"+
		"  Payee: %s",
		issue.Coin.String(),
		issue.Payee.String(),
	)
}

// NewIssue creates a new Issue object.
func NewIssue(coin sdk.Coin, payee sdk.AccAddress) Issue {
	return Issue{
		Coin:  coin,
		Payee: payee,
	}
}
