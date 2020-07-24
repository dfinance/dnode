package types

import (
	"bytes"

	dnTypes "github.com/dfinance/dnode/helpers/types"
)

var (
	KeyDelimiter    = []byte(":")
	KeyMarketPrefix = []byte("market")
	KeyLastMarketId = []byte("last_market_id")
)

// GetMarketsKey returns key for storing markets.
func GetMarketsKey(id dnTypes.ID) []byte {
	return bytes.Join(
		[][]byte{
			KeyMarketPrefix,
			[]byte(id.String()),
		},
		KeyDelimiter,
	)
}

// GetPrefixMarketsKey return storage key prefix for markets (used for iteration).
func GetPrefixMarketsKey() []byte {
	return append(KeyMarketPrefix, KeyDelimiter...)
}
