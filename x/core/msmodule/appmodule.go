package msmodule

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
)

// AppMsModule extends std AppModule interface with multi signature handler getter.
type AppMsModule interface {
	module.AppModule
	NewMsHandler() MsHandler
}

// MsHandler is a multi signature message handler.
type MsHandler func(ctx sdk.Context, msg MsMsg) error

// MsMsg defines multi signature message interface.
// Used to execute message call, once call is confirmed.
type MsMsg interface {
	Route() string
	Type() string
	ValidateBasic() error
}
