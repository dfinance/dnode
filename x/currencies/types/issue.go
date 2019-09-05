package types

import (
    sdk "github.com/cosmos/cosmos-sdk/types"
    "fmt"
)

type Issue struct {
    CurrencyId sdk.Int        `json:"currencyId"`
    Symbol     string         `json:"symbol"`
    Amount     sdk.Int        `json:"amount"`
    Recipient  sdk.AccAddress `json:"recipient"`
}

func NewIssue(currencyId sdk.Int, symbol string, amount sdk.Int, recipient sdk.AccAddress) Issue {
    return Issue{
        CurrencyId: currencyId,
        Symbol:     symbol,
        Amount:     amount,
        Recipient:  recipient,
    }
}

func (issue Issue) String() string {
    return fmt.Sprintf("Issue: \n" +
        "\tCurrency Id: %s\n" +
        "\tSymbol:      %s\n" +
        "\tAmount:      %s\n" +
        "\tRecipient:   %s\n",
        issue.CurrencyId, issue.Symbol,
        issue.Amount.String(), issue.Recipient.String())
}
