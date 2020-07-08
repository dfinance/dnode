package types

import (
	"bytes"

	dnTypes "github.com/dfinance/dnode/helpers/types"
)

const (
	ModuleName   = "currencies"
	RouterKey    = ModuleName
	StoreKey     = ModuleName
	GovRouterKey = RouterKey
)

var (
	KeyDelimiter = []byte(":")
)

// GetIssuesKey returns key for storing issues.
func GetIssuesKey(id string) []byte {
	return bytes.Join(
		[][]byte{
			[]byte("issue"),
			[]byte(id),
		},
		KeyDelimiter,
	)
}

// GetWithdrawKey returns key for storing withdraw.
func GetWithdrawKey(id dnTypes.ID) []byte {
	return bytes.Join(
		[][]byte{
			[]byte("withdraw"),
			[]byte(id.String()),
		},
		KeyDelimiter,
	)
}

// GetLastWithdrawIDKey returns storage key for withdrawID.
func GetLastWithdrawIDKey() []byte {
	return []byte("lastWithdrawID")
}
