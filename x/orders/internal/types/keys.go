package types

import (
	"bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"

	dnTypes "github.com/dfinance/dnode/helpers/types"
)

// Storage keys.
var (
	KeyDelimiter   = []byte(":")
	OrderKeyPrefix = []byte("order")
	LastOrderIDKey = []byte("last_order_id")
)

// GetOrderKey returns storage key for order ID.
func GetOrderKey(id dnTypes.ID) []byte {
	return bytes.Join(
		[][]byte{
			OrderKeyPrefix,
			sdk.Uint64ToBigEndian(id.UInt64()),
		},
		KeyDelimiter,
	)
}
