package types

import "fmt"

const (
	// ModuleKey is the name of the module
	ModuleName = "pricefeed"

	// StoreKey is the store key string for gov
	StoreKey = ModuleName

	// RouterKey is the message route for gov
	RouterKey = ModuleName

	// QuerierRoute is the querier route for gov
	QuerierRoute = ModuleName

	// Parameter store default namestore
	DefaultParamspace = ModuleName

	// Store prefix for the raw price feed of an asset
	RawPriceFeedPrefix = StoreKey + ":raw:"

	// Store prefix for the current price of an asset
	CurrentPricePrefix = StoreKey + ":currentprice:"

	// Store Prefix for the assets in the price feed system
	AssetPrefix = StoreKey + ":assets"

	// OraclePrefix store prefix for the price feed accounts
	OraclePrefix = StoreKey + ":pricefeeds"
)

// Get a key to store PostedPrices for specific assetCode and blockHeight
func GetRawPricesKey(assetCode string, blockHeight int64) []byte {
	return []byte(fmt.Sprintf("%s%s:%d", RawPriceFeedPrefix, assetCode, blockHeight))
}
