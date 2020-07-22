package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	dnTypes "github.com/dfinance/dnode/helpers/types"
)

const (
	QueryCurrency   = "currency"
	QueryCurrencies = "currencies"
	QueryIssue      = "issue"
	QueryWithdraws  = "withdraws"
	QueryWithdraw   = "withdraw"
)

// Client request for currency.
type CurrencyReq struct {
	Denom string `json:"denom" yaml:"denom"`
}

// Client request for issue.
type IssueReq struct {
	ID string `json:"id" yaml:"id"`
}

// Client request for withdraw.
type WithdrawReq struct {
	ID dnTypes.ID `json:"id" yaml:"id"`
}

// Client request for withdraws.
type WithdrawsReq struct {
	Page  sdk.Uint `json:"page" yaml:"page"`
	Limit sdk.Uint `json:"limit" yaml:"limit"`
}
