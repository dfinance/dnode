package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	dnTypes "github.com/dfinance/dnode/helpers/types"
)

const (
	QueryCurrency  = "currency"
	QueryIssue     = "issue"
	QueryWithdraws = "withdraws"
	QueryWithdraw  = "withdraw"
)

// Client request for currency.
type CurrencyReq struct {
	Denom string
}

// Client request for issue.
type IssueReq struct {
	ID string
}

// Client request for withdraw.
type WithdrawReq struct {
	ID dnTypes.ID
}

// Client request for withdraws.
type WithdrawsReq struct {
	Page  sdk.Uint
	Limit sdk.Uint
}
