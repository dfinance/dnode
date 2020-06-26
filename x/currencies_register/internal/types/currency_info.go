package types

import (
	"encoding/hex"
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"

	dnTypes "github.com/dfinance/dnode/helpers/types"
	"github.com/dfinance/dnode/x/common_vm"
)

// Contains currency info.
type CurrencyInfo struct {
	Denom       []byte   `json:"denom" swaggertype:"string" example:"dfi"`
	Decimals    uint8    `json:"decimals"`
	// If true, currency created in DVM using 0x1::Dfinance::tokenize func.
	IsToken     bool     `json:"isToken"`
	// Owner is 0x1 for non-token currency and account address for token currencies.
	Owner       []byte   `json:"owner" lcs:"len=20" swaggertype:"string"`
	TotalSupply *big.Int `json:"totalSupply"`
}

// New currency.
func NewCurrencyInfo(denom []byte, decimals uint8, isToken bool, owner []byte, totalSupply sdk.Int) (CurrencyInfo, error) {
	if err := dnTypes.DenomFilter(string(denom)); err != nil {
		return CurrencyInfo{}, fmt.Errorf("denom: %w", err)
	}

	if len(owner) != common_vm.VMAddressLength {
		return CurrencyInfo{}, fmt.Errorf("owner: address length is not equal to VM address length: %d / %d", len(owner), common_vm.VMAddressLength)
	}

	if totalSupply.IsNegative() {
		return CurrencyInfo{}, fmt.Errorf("totalSupply: negative")
	}

	return CurrencyInfo{
		Denom:       denom,
		Decimals:    decimals,
		IsToken:     isToken,
		Owner:       owner,
		TotalSupply: totalSupply.BigInt(),
	}, nil
}

// UintToDec converts sdk.Uint to sdk.Dec using currency decimals.
func (c CurrencyInfo) UintToDec(quantity sdk.Uint) sdk.Dec {
	return sdk.NewDecFromIntWithPrec(sdk.Int(quantity), int64(c.Decimals))
}

// DecToUint converts sdk.Dec to sdk.Uint using currency decimals.
func (c CurrencyInfo) DecToUint(quantity sdk.Dec) sdk.Uint {
	res := quantity.Quo(c.MinDecimal()).TruncateInt()

	return sdk.NewUintFromBigInt(res.BigInt())
}

// MinDecimal return minimal currency value.
func (c CurrencyInfo) MinDecimal() sdk.Dec {
	return sdk.NewDecFromIntWithPrec(sdk.OneInt(), int64(c.Decimals))
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
