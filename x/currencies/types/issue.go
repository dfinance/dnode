// Issue type implementation for currencies.
package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Issue struct {
	Symbol    string         `json:"symbol" example:"dfi"` // Denom
	Amount    sdk.Int        `json:"amount" example:"100"`
	Recipient sdk.AccAddress `json:"recipient" example:"wallet13jyjuz3kkdvqw8u4qfkwd94emdl3vx394kn07h"`
}

func NewIssue(symbol string, amount sdk.Int, recipient sdk.AccAddress) Issue {
	return Issue{
		Symbol:    symbol,
		Amount:    amount,
		Recipient: recipient,
	}
}

func (issue Issue) String() string {
	return fmt.Sprintf("Issue: \n"+
		"\tSymbol:      %s\n"+
		"\tAmount:      %s\n"+
		"\tRecipient:   %s\n",
		issue.Symbol, issue.Amount.String(), issue.Recipient.String())
}
