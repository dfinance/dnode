package types

import (
	"encoding/hex"
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dfinance/dnode/x/common_vm"
)

// Contains currency info.
type CurrencyInfo struct {
	Denom       []byte   `json:"denom" swaggertype:"string" example:"dfi"`
	Decimals    uint8    `json:"decimals"`
	IsToken     bool     `json:"isToken"`
	Owner       []byte   `json:"owner" lcs:"len=20" swaggertype:"string" example:"dfi"`
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
