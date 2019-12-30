package types

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/types"
)

type ValidatorsConfirmations struct {
	Validators    Validators `json:"validators"`
	Confirmations uint16     `json:"confirmations"`
}

func (q ValidatorsConfirmations) String() string {
	return fmt.Sprintf("%v\n"+
		"Confirmations: %d",
		q.Validators, q.Confirmations)
}

type QueryValidator struct {
	Address types.AccAddress
}
