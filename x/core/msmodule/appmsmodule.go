// Implements AppMsModule interface (inherits from AppModule) to manage also multisignature modules.
package msmodule

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
)

// Base message interface, using to execute by call, once call confirmed
type MsMsg interface {
	Route() string
	Type() string
	ValidateBasic() error
}

// Multisignature handler.
type MsHandler func(ctx sdk.Context, msg MsMsg) error

// Message handle for multisignature calls
type AppMsModule interface {
	module.AppModule
	NewMsHandler() MsHandler
}
