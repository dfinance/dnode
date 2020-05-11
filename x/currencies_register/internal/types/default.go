package types

import (
	"fmt"
)

const (
	ModuleName = "currencies_register"
	StoreKey   = ModuleName
)

// Get currency path key.
func GetCurrencyPathKey(denom string) []byte {
	return []byte(fmt.Sprintf("currency_path:%s", denom))
}
