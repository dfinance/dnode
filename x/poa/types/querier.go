// Types for querier.
package types

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/types"
)

// Response from querier with confirmations and validators list.
type ValidatorsConfirmations struct {
	Validators    Validators `json:"validators"`                // Registered validators list
	Confirmations uint16     `json:"confirmations" example:"3"` // Minimum number of confirmations needed to approve Call
}

func (q ValidatorsConfirmations) String() string {
	return fmt.Sprintf("%v\n"+
		"Confirmations: %d",
		q.Validators, q.Confirmations)
}

// Request for querier to export validators by address.
type QueryValidator struct {
	Address types.AccAddress
}
