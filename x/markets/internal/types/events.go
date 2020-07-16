package types

import sdk "github.com/cosmos/cosmos-sdk/types"

const (
	EventTypeCreate = ModuleName + ".create"
	//
	AttributeMarketId   = "market_id"
	AttributeBaseDenom  = "base_denom"
	AttributeQuoteDenom = "quote_denom"
)

// NewMarketCreatedEvent creates an Event on market creation.
func NewMarketCreatedEvent(market Market) sdk.Event {
	return sdk.NewEvent(
		EventTypeCreate,
		sdk.NewAttribute(AttributeMarketId, market.ID.String()),
		sdk.NewAttribute(AttributeBaseDenom, market.BaseAssetDenom),
		sdk.NewAttribute(AttributeQuoteDenom, market.QuoteAssetDenom),
	)
}
