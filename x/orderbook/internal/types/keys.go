package types

import (
	"bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"

	dnTypes "github.com/dfinance/dnode/helpers/types"
)

// Storage keys.
var (
	KeyDelimiter         = []byte(":")
	HistoryItemKeyPrefix = []byte("history_item")
)

// GetOrderKey returns storage key for order ID.
func GetHistoryItemKey(marketID dnTypes.ID, blockHeight int64) []byte {
	return bytes.Join(
		[][]byte{
			HistoryItemKeyPrefix,
			sdk.Uint64ToBigEndian(marketID.UInt64()),
			sdk.Uint64ToBigEndian(uint64(blockHeight)),
		},
		KeyDelimiter,
	)
}
