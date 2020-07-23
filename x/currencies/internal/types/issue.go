package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	dnTypes "github.com/dfinance/dnode/helpers/types"
)

// Issue is an info about issuing currency to the payee (recipient).
type Issue struct {
	// Issuing coin
	Coin sdk.Coin `json:"coin" yaml:"coin" swaggertype:"string" example:"100dfi"`
	// Target account for increasing coin balance
	Payee sdk.AccAddress `json:"payee" yaml:"payee" swaggertype:"string" format:"bech32" example:"wallet13jyjuz3kkdvqw8u4qfkwd94emdl3vx394kn07h"`
}

// Valid checks that issue is valid (used for genesis ops).
func (issue Issue) Valid() error {
	if err := dnTypes.DenomFilter(issue.Coin.Denom); err != nil {
		return fmt.Errorf("coin: denom: %w", err)
	}
	if issue.Coin.Amount.LTE(sdk.ZeroInt()) {
		return fmt.Errorf("coin: amount: LTE to zero")
	}

	if issue.Payee.Empty() {
		return fmt.Errorf("payee: empty")
	}

	return nil
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
