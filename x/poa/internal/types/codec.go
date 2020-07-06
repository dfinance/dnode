package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

var ModuleCdc *codec.Codec

// RegisterCodec registers module specific messages.
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgAddValidator{}, ModuleName+"/AddValidator", nil)
	cdc.RegisterConcrete(MsgRemoveValidator{}, ModuleName+"/RemoveValidator", nil)
	cdc.RegisterConcrete(MsgReplaceValidator{}, ModuleName+"/ReplaceValidator", nil)
}

func init() {
	cdc := codec.New()
	RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	ModuleCdc = cdc.Seal()
}
