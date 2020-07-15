package types

import (
	"bytes"
	"strconv"

	"github.com/dfinance/dnode/helpers/types"
)

const (
	ModuleName        = "oracle"
	StoreKey          = ModuleName
	RouterKey         = ModuleName
	DefaultParamspace = ModuleName
)

var (
	KeyDelimiter    = []byte(":")
	ModuleKey       = []byte(ModuleName)
	RawPriceKey     = []byte("raw")
	CurrentPriceKey = []byte("currentprice")
)

// GetRawPricesKey Get a key to store PostedPrices for specific assetCode and blockHeight.
func GetRawPricesKey(assetCode types.AssetCode, blockHeight int64) []byte {
	return bytes.Join(
		[][]byte{
			ModuleKey,
			RawPriceKey,
			[]byte(assetCode),
			[]byte(strconv.FormatInt(blockHeight, 10)),
		},
		KeyDelimiter,
	)
}

// GetCurrentPriceKey Get a key to store CurrentPrice for specific assetCode.
func GetCurrentPriceKey(assetCode types.AssetCode) []byte {
	return bytes.Join(
		[][]byte{
			ModuleKey,
			CurrentPriceKey,
			[]byte(assetCode),
		},
		KeyDelimiter,
	)
}
