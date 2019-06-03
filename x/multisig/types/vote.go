package types

import "github.com/cosmos/cosmos-sdk/types"

type Vote struct {
	Address types.AccAddress
}

type Votes []Vote