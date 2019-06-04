package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Base message interface, using to execute by call, once call confirmed
type MsMsg interface {
	Route() 		string
	Type() 		    string
	ValidateBasic() sdk.Error
	GetSignBytes()  []byte
	GetSigners()	[]sdk.AccAddress
}