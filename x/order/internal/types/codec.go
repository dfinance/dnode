package types

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
)

var ModuleCdc = codec.New()

func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgPostOrder{}, fmt.Sprintf("%s/%s", ModuleName, MsgPostOrder{}.Type()), nil)
	cdc.RegisterConcrete(MsgRevokeOrder{}, fmt.Sprintf("%s/%s", ModuleName, MsgRevokeOrder{}.Type()), nil)
}
