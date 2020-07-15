package types

import (
	"fmt"

	"github.com/dfinance/dnode/helpers/types"
)

const (
	ModuleName         = "oracle"
	StoreKey           = ModuleName
	RouterKey          = ModuleName
	DefaultParamspace  = ModuleName
	//
	RawPriceFeedPrefix = StoreKey + ":raw:"          // Store prefix for the raw oracle of an asset
	CurrentPricePrefix = StoreKey + ":currentprice:" // Store prefix for the current price of an asset
	AssetPrefix        = StoreKey + ":assets"        // Store Prefix for the assets in the oracle system
	OraclePrefix       = StoreKey + ":oracles"       // OraclePrefix store prefix for the oracle accounts
)

// GetRawPricesKey Get a key to store PostedPrices for specific assetCode and blockHeight.
func GetRawPricesKey(assetCode types.AssetCode, blockHeight int64) []byte {
	return []byte(fmt.Sprintf("%s%s:%d", RawPriceFeedPrefix, assetCode.String(), blockHeight))
}
