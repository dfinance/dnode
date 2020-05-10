package types

import (
	"encoding/hex"
	"fmt"
)

// Currency path structure, contains libra path for specific currency info.
type CurrencyPath struct {
	Path []byte `json:"path"`
}

// Create new currency path.
func NewCurrencyPath(path []byte) CurrencyPath {
	return CurrencyPath{Path: path}
}

// Currency path to string.
func (c CurrencyPath) String() string {
	return fmt.Sprintf("Path: %s", hex.EncodeToString(c.Path))
}
