package types

import (
    sdk "github.com/cosmos/cosmos-sdk/types"
    "fmt"
)

type Issue struct {
    Symbol    string         `json:"symbol"`
    Amount    sdk.Int        `json:"amount"`
    Recipient sdk.AccAddress `json:"recipient"`
}

func NewIssue(symbol string, amount sdk.Int, recipient sdk.AccAddress) Issue {
    return Issue{
        Symbol:    symbol,
        Amount:    amount,
        Recipient: recipient,
    }
}

func (issue Issue) String() string {
    return fmt.Sprintf("Issue: \n" +
        "\tSymbol:    %s\n" +
        "\tAmount:    %s\n" +
        "\tRecipient: %s\n",
        issue.Symbol, issue.Amount.String(),
        issue.Recipient.String())
}
