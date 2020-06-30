package types

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/dfinance/dnode/x/common_vm"
)

// CurrencyInfo contains Currency meta-data used as a WriteSet by VM.
// For standard currencies, CurrencyInfo is converted from Currency.
// For token currencies, CurrencyInfo is created by VM.
type CurrencyInfo struct {
	// Currency denom ([]byte is used for VM)
	Denom []byte `json:"denom" swaggertype:"string" example:"dfi"`
	// Number of currency decimals
	Decimals uint8 `json:"decimals"`
	// If true, currency is created by DVM using 0x1::Dfinance::tokenize func
	IsToken bool `json:"isToken"`
	// Owner is 0x1 for non-token currency and account address for token currencies
	Owner []byte `json:"owner" lcs:"len=20" swaggertype:"string"`
	// Total amount of currency coins in Bank
	TotalSupply *big.Int `json:"totalSupply"`
}

func (c CurrencyInfo) String() string {
	return fmt.Sprintf("CurrencyInfo:\n"+
		"  Denom:    %s\n"+
		"  Decimals: %d\n"+
		"  Is Token: %t\n"+
		"  Owner:    %s\n"+
		"  Total supply: %s",
		string(c.Denom),
		c.Decimals,
		c.IsToken,
		hex.EncodeToString(c.Owner),
		c.TotalSupply.String(),
	)
}

// NewCurrencyInfo converts Currency to VM's CurrencyInfo checking if owner is stdlib.
// Contract: Currency object is expected to be valid.
func NewCurrencyInfo(currency Currency, ownerAddress []byte) (CurrencyInfo, error) {
	if len(ownerAddress) != common_vm.VMAddressLength {
		return CurrencyInfo{}, fmt.Errorf("ownerAddress: address length is not equal to VM address length: %d / %d", len(ownerAddress), common_vm.VMAddressLength)
	}

	isToken := false
	if bytes.Compare(ownerAddress, common_vm.StdLibAddress) != 0 {
		isToken = true
	}

	return CurrencyInfo{
		Denom:       []byte(currency.Denom),
		Decimals:    currency.Decimals,
		IsToken:     isToken,
		Owner:       ownerAddress,
		TotalSupply: currency.Supply.BigInt(),
	}, nil
}
