package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Base message interface, using to execute by call, once call confirmed
type MsMsg interface {
	Route() 		string
	Type() 		    string
	ValidateBasic() sdk.Error
}

// Message handle for multisignature calls
type MsHandler func(ctx sdk.Context, msg MsMsg) sdk.Error