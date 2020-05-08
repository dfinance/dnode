package types

import (
	"encoding/hex"
	"fmt"
)

type CurrencyPath struct {
	Path []byte `json:"path"`
}

func NewCurrencyPath(path []byte) CurrencyPath {
	return CurrencyPath{Path: path}
}

func (c CurrencyPath) String() string {
	return fmt.Sprintf("Currency path: %s", hex.EncodeToString(c.Path))
}
