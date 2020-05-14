package types

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
)

var ModuleCdc = codec.New()

// RegisterCodec registers module specific messages.
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgCreateMarket{}, fmt.Sprintf("%s/%s", ModuleName, MsgCreateMarket{}.Type()), nil)
}
