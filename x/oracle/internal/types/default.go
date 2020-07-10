package types

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"strings"
)

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

// GetRawPricesKey Get a key to store PostedPrices for specific assetCode and blockHeight.
func GetRawPricesKey(assetCode string, blockHeight int64) []byte {
	return []byte(fmt.Sprintf("%s%s:%d", RawPriceFeedPrefix, assetCode, blockHeight))
}

// ParseOracles parses coma-separate notation oracle addresses and returns Oracles object.
func ParseOracles(addresses string) (Oracles, error) {
	res := make([]Oracle, 0)
	for _, address := range strings.Split(addresses, ",") {
		address = strings.TrimSpace(address)
		if len(address) == 0 {
			continue
		}
		oracleAddress, err := ValidateAddress(address)
		if err != nil {
			return nil, err
		}

		oracle := NewOracle(oracleAddress)

		res = append(res, oracle)
	}

	return res, nil
}

// ValidateAddress validates the oracle string-notation address.
func ValidateAddress(address string) (sdk.AccAddress, error) {
	oracle, err := sdk.AccAddressFromBech32(address)
	if err != nil {
		return nil, err
	}

	return oracle, nil
}
