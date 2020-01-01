package types

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Described validator.
type Validator struct {
	Address    sdk.AccAddress `json:"address"`
	EthAddress string         `json:"eth_address"`
}

// Array of validators.
type Validators []Validator

// Creating new validator instance.
func NewValidator(address sdk.AccAddress, ethAddress string) Validator {
	return Validator{
		Address:    address,
		EthAddress: ethAddress,
	}
}

func (v Validator) String() string {
	return fmt.Sprintf("Address: %s\n"+
		"Eth Address %s", v.Address.String(), v.EthAddress)
}
