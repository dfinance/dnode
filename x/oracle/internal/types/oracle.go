package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Oracle struct that documents which address an oracle is using.
type Oracle struct {
	// Address
	Address sdk.AccAddress `json:"address" yaml:"address"`
}

// String implements fmt.Stringer.
func (o Oracle) String() string {
	return fmt.Sprintf(`Address: %s`, o.Address)
}

// NewOracle creates a new Oracle.
func NewOracle(address sdk.AccAddress) Oracle {
	return Oracle{
		Address: address,
	}
}

// Oracles array type for oracle.
type Oracles []Oracle

// String implements fmt.Stringer.
func (os Oracles) String() string {
	out := "Oracles:\n"
	for _, o := range os {
		out += fmt.Sprintf("%s\n", o.String())
	}

	return strings.TrimSpace(out)
}
