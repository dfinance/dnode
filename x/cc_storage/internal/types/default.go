package types

import (
	"bytes"
)

const (
	ModuleName        = "cc_storage"
	StoreKey          = ModuleName
	DefaultParamspace = ModuleName
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
