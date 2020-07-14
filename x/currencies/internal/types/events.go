package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	dnTypes "github.com/dfinance/dnode/helpers/types"
)

const (
	EventTypesIssue    = ModuleName + ".issue"
	EventTypesWithdraw = ModuleName + ".withdraw"
	//
	AttributeDenom      = "denom"
	AttributeAmount     = "amount"
	AttributeIssueId    = "issue_id"
	AttributeWithdrawId = "withdraw_id"
	AttributeSender     = "sender"
)

// NewIssueEvent creates an Event on currency issue.
func NewIssueEvent(id string, coin sdk.Coin, payee sdk.AccAddress) sdk.Event {
	return sdk.NewEvent(
		EventTypesIssue,
		sdk.NewAttribute(AttributeIssueId, id),
		sdk.NewAttribute(AttributeDenom, coin.Denom),
		sdk.NewAttribute(AttributeAmount, coin.Amount.String()),
		sdk.NewAttribute(AttributeSender, payee.String()),
	)
}

// NewWithdrawEvent creates an Event on currency withdraw.
func NewWithdrawEvent(id dnTypes.ID, coin sdk.Coin, spender sdk.AccAddress) sdk.Event {
	return sdk.NewEvent(
		EventTypesWithdraw,
		sdk.NewAttribute(AttributeWithdrawId, id.String()),
		sdk.NewAttribute(AttributeDenom, coin.Denom),
		sdk.NewAttribute(AttributeAmount, coin.Amount.String()),
		sdk.NewAttribute(AttributeSender, spender.String()),
	)
}
