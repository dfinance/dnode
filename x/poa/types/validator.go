package types

import sdk "github.com/cosmos/cosmos-sdk/types"

type Validator struct {
	address    sdk.AccAddress	`json:"address"`
	ethAddress string			`json:"eth_address"`
}

func NewValidator(address sdk.AccAddress, ethAddress string) Validator {
	return Validator{
		address: 	address,
		ethAddress: ethAddress,
	}
}