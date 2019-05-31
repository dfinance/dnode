package types

import sdk "github.com/cosmos/cosmos-sdk/types"

type Validator struct {
	Address    sdk.AccAddress	`json:"address"`
	EthAddress string			`json:"eth_address"`
}

func NewValidator(address sdk.AccAddress, ethAddress string) Validator {
	return Validator{
		Address: 	address,
		EthAddress: ethAddress,
	}
}