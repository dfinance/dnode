package types

import (
	"fmt"
	"strings"
)

type AssetCode string

// Validate validates asset code.
func (a AssetCode) Validate() error {
	return AssetCodeFilter(a.String())
}

// String returns string enum representation.
func (a AssetCode) String() string {
	return string(a)
}

// ReverseCode reverses asset code.
// Will panic if use it with the wrong format asset code.
func (a AssetCode) ReverseCode() AssetCode {
	parts := strings.Split(a.String(), "_")
	if len(parts) != 2 {
		panic(fmt.Errorf("wrong asset code format: %s", a))
	}
	return AssetCode(parts[1] + "_" + parts[0])
}
