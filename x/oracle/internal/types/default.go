package types

import "fmt"

const (
	ModuleName         = "oracle"                    // ModuleKey is the name of the module
	StoreKey           = ModuleName                  // StoreKey is the store key string for gov
	RouterKey          = ModuleName                  // RouterKey is the message route for gov
	QuerierRoute       = ModuleName                  // QuerierRoute is the querier route for gov
	DefaultParamspace  = ModuleName                  // Parameter store default namestore
	RawPriceFeedPrefix = StoreKey + ":raw:"          // Store prefix for the raw oracle of an asset
	CurrentPricePrefix = StoreKey + ":currentprice:" // Store prefix for the current price of an asset
	AssetPrefix        = StoreKey + ":assets"        // Store Prefix for the assets in the oracle system
	OraclePrefix       = StoreKey + ":oracles"       // OraclePrefix store prefix for the oracle accounts
)

// Get a key to store PostedPrices for specific assetCode and blockHeight
func GetRawPricesKey(assetCode string, blockHeight int64) []byte {
	return []byte(fmt.Sprintf("%s%s:%d", RawPriceFeedPrefix, assetCode, blockHeight))
}
