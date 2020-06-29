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

// GetDestroyKey returns key for storing destroy.
func GetDestroyKey(id dnTypes.ID) []byte {
	return bytes.Join(
		[][]byte{
			[]byte("destroy"),
			[]byte(id.String()),
		},
		KeyDelimiter,
	)
}

// GetLastDestroyIDKey returns storage key for destroyID.
func GetLastDestroyIDKey() []byte {
	return []byte("lastDestroyID")
}
