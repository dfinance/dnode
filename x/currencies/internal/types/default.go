package types

import (
	"bytes"

	dnTypes "github.com/dfinance/dnode/helpers/types"
)

const (
	ModuleName        = "currencies"
	RouterKey         = ModuleName
	StoreKey          = ModuleName
	DefaultParamspace = ModuleName
	GovRouterKey      = RouterKey
)

var (
	KeyDelimiter      = []byte(":")
	KeyCurrencyPrefix = []byte("currency")
)

// GetCurrencyKey returns Key for storing currency.
func GetCurrencyKey(denom string) []byte {
	return bytes.Join(
		[][]byte{
			KeyCurrencyPrefix,
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

// GetCurrencyBalancePathKey returns storage key for currencyBalance VM path.
func GetCurrencyBalancePathKey(denom string) []byte {
	return bytes.Join(
		[][]byte{
			[]byte("currencyBalancePath"),
			[]byte(denom),
		},
		KeyDelimiter,
	)
}

// GetCurrencyInfoPathKey returns storage key for currencyInfo VM path.
func GetCurrencyInfoPathKey(denom string) []byte {
	return bytes.Join(
		[][]byte{
			[]byte("currencyInfoPath"),
			[]byte(denom),
		},
		KeyDelimiter,
	)
}
