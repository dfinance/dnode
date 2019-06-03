package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type MsMsg interface {
	Route() string
	Type() string
	ValidateBasic() sdk.Error
}

// Handler defines a function that handles a proposal after it has passed the
// governance process.
type Handler func(ctx sdk.Context, call Call) sdk.Error