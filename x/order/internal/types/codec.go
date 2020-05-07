package types

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
)

var ModuleCdc = codec.New()

func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgPostOrder{}, fmt.Sprintf("%s/%s", ModuleName, MsgPostOrder{}.Type()), nil)
	cdc.RegisterConcrete(MsgCancelOrder{}, fmt.Sprintf("%s/%s", ModuleName, MsgCancelOrder{}.Type()), nil)
}
