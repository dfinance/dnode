package types

import "github.com/cosmos/cosmos-sdk/types"

// Base vote that contains address of validator
type Vote struct {
	Address types.AccAddress
}

// Votes slice
type Votes []Vote