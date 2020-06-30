package types

import (
	"bytes"

	dnTypes "github.com/dfinance/dnode/helpers/types"
)

const (
	ModuleName = "currencies"
	RouterKey  = ModuleName
	StoreKey   = ModuleName
)

var (
	KeyDelimiter = []byte(":")
)

// GetCurrencyKey returns Key for storing currency.
func GetCurrencyKey(denom string) []byte {
	return bytes.Join(
		[][]byte{
			[]byte("currency"),
			[]byte(denom),
		},
		KeyDelimiter,
	)
}

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
