package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"fmt"
)

type Validator struct {
	Address    sdk.AccAddress	`json:"address"`
	EthAddress string			`json:"eth_address"`
}

type Validators []Validator

func NewValidator(address sdk.AccAddress, ethAddress string) Validator {
	return Validator{
		Address: 	address,
		EthAddress: ethAddress,
	}
}

func (v Validator) String() string {
	return fmt.Sprintf("Address: %s\n" +
		"Eth Address %s", v.Address.String(), v.EthAddress)
}