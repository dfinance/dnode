package types

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/dfinance/dnode/x/common_vm"
)

// ResCurrencyInfo is a DVM resource, containing Currency meta-data.
// For standard currencies, CurrencyInfo is converted from Currency.
// For token currencies, CurrencyInfo is created by VM.
type ResCurrencyInfo struct {
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

func (c ResCurrencyInfo) String() string {
	return fmt.Sprintf("ResCurrencyInfo:\n"+
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

// NewResCurrencyInfo converts Currency to VM's ResCurrencyInfo checking if owner is stdlib.
// Contract: Currency object is expected to be valid.
func NewResCurrencyInfo(currency Currency, ownerAddress []byte) (ResCurrencyInfo, error) {
	if len(ownerAddress) != common_vm.VMAddressLength {
		return ResCurrencyInfo{}, fmt.Errorf("ownerAddress: address length is not equal to VM address length: %d / %d", len(ownerAddress), common_vm.VMAddressLength)
	}

	isToken := false
	if !bytes.Equal(ownerAddress, common_vm.StdLibAddress) {
		isToken = true
	}

	return ResCurrencyInfo{
		Denom:       []byte(currency.Denom),
		Decimals:    currency.Decimals,
		IsToken:     isToken,
		Owner:       ownerAddress,
		TotalSupply: currency.Supply.BigInt(),
	}, nil
}
