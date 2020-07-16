package types

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	EventTypesCreate = ModuleName + ".create"
	//
	AttributeDenom    = "denom"
	AttributeDecimals = "decimals"
	AttributeInfoPath = "info_path"
)

// NewCCCreatedEvent creates an Event on currency creation.
func NewCCCreatedEvent(currency Currency, params CurrencyParams) sdk.Event {
	return sdk.NewEvent(
		EventTypesCreate,
		sdk.NewAttribute(AttributeDenom, currency.Denom),
		sdk.NewAttribute(AttributeDecimals, strconv.FormatUint(uint64(currency.Decimals), 10)),
		sdk.NewAttribute(AttributeInfoPath, params.InfoPathHex),
	)
}
