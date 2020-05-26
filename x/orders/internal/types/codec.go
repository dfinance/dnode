package types

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
)

var ModuleCdc *codec.Codec

// RegisterCodec registers module specific messages.
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgPostOrder{}, fmt.Sprintf("%s/MsgPostOrder", ModuleName), nil)
	cdc.RegisterConcrete(MsgRevokeOrder{}, fmt.Sprintf("%s/MsgRevokeOrder", ModuleName), nil)
}

func init() {
	cdc := codec.New()
	RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	ModuleCdc = cdc.Seal()
}
