package queries

import (
	"fmt"
	poaTypes "wings-blockchain/x/poa/types"
)

// Get validators response
type QueryValidatorsRes struct {
	Validators    poaTypes.Validators `json:"validators"`
	Amount     	  int			   	  `json:"amount"`
	Confirmations int 			   	  `json:"confirmations"`
}

func (q QueryValidatorsRes) String() string {
	return fmt.Sprintf("%v\n" +
		"Amount: %d\n" +
		"Confirmations: %d",
		q.Validators, q.Amount, q.Confirmations)
}
