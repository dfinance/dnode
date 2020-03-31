package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Described validator.
type Validator struct {
	Address    sdk.AccAddress `json:"address" example:"wallet1a7280dyzp487r7wghr99f6r3h2h2z4gk4d740m"`
	EthAddress string         `json:"eth_address" example:"0x29D7d1dd5B6f9C864d9db560D72a247c178aE86B"`
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
