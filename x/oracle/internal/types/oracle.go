package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Oracle struct contains oracle source meta.
type Oracle struct {
	// Address
	Address sdk.AccAddress `json:"address" yaml:"address" swaggertype:"string" format:"bech32" example:"wallet13jyjuz3kkdvqw8u4qfkwd94emdl3vx394kn07h"`
}

func (o Oracle) String() string {
	return fmt.Sprintf(`Address: %s`, o.Address)
}

// NewOracle creates a new Oracle.
func NewOracle(address sdk.AccAddress) Oracle {
	return Oracle{
		Address: address,
	}
}

// Oracles slice type for oracle.
type Oracles []Oracle

func (list Oracles) String() string {
	strBuilder := strings.Builder{}

	strBuilder.WriteString("Oracles:\n")
	for i, oracle := range list {
		strBuilder.WriteString(oracle.String())
		if i < len(list) - 1 {
			strBuilder.WriteString("\n")
		}
	}

	return strBuilder.String()
}
