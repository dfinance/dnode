// Implements AppMsModule interface (inherits from AppModule) to manage also multisignature modules.
package core

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
)

// Base message interface, using to execute by call, once call confirmed
type MsMsg interface {
	Route() string
	Type() string
	ValidateBasic() sdk.Error
}

// Multisignature handler.
type MsHandler func(ctx sdk.Context, msg MsMsg) sdk.Error

// Message handle for multisignature calls
type AppMsModule interface {
	module.AppModule
	NewMsHandler() MsHandler
}
