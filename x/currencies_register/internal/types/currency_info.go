package types

import (
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/dfinance/dnode/x/common_vm"
)

// Contains currency info.
type CurrencyInfo struct {
	Denom       []byte   `json:"denom"`
	Decimals    uint8    `json:"decimals"`
	IsToken     bool     `json:"isToken"`
	Owner       []byte   `json:"owner" lcs:"len=20"`
	TotalSupply *big.Int `json:"totalSupply"`
}

// New currency.
func NewCurrencyInfo(denom []byte, decimals uint8, isToken bool, owner []byte, totalSupply *big.Int) (CurrencyInfo, error) {
	if len(owner) != common_vm.VMAddressLength {
		return CurrencyInfo{}, fmt.Errorf("length of owner address is not equal to address length: %d / %d", len(owner), common_vm.VMAddressLength)
	}

	return CurrencyInfo{
		Denom:       denom,
		Decimals:    decimals,
		IsToken:     isToken,
		Owner:       owner,
		TotalSupply: totalSupply,
	}, nil
}

// Currency to string.
func (c CurrencyInfo) String() string {
	return fmt.Sprintf("Currency: %s\n"+
		"\tDecimals: %d\n"+
		"\tIs Token: %t\n"+
		"\tOwner:    %s\n"+
		"\tTotal supply: %s\n",
		string(c.Denom), c.Decimals, c.IsToken,
		hex.EncodeToString(c.Owner), c.TotalSupply.String())
}
