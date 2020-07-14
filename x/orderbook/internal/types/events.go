package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	EventTypeClearance = ModuleName + ".clearance"
	//
	AttributeMarketId = "market_id"
	AttributePrice    = "price"
)

// NewClearanceEvent creates an Event on successful market match.
func NewClearanceEvent(result MatcherResult) sdk.Event {
	return sdk.NewEvent(
		EventTypeClearance,
		sdk.NewAttribute(AttributeMarketId, result.MarketID.String()),
		sdk.NewAttribute(AttributePrice, result.ClearanceState.Price.String()),
	)
}
