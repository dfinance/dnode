package poa

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"wings-blockchain/x/poa/msgs"
)

// Registering amino types for PoA messages
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(msgs.MsgAddValidator{}, msgs.MsgAddValidatorType, nil)
	cdc.RegisterConcrete(msgs.MsgRemoveValidator{}, msgs.MsgRemoveValidatorType, nil)
	cdc.RegisterConcrete(msgs.MsgReplaceValidator{}, msgs.MsgReplaceValidatorType, nil)
}

// module wide codec
var ModuleCdc *codec.Codec

func init() {
	ModuleCdc = codec.New()
	RegisterCodec(ModuleCdc)
	codec.RegisterCrypto(ModuleCdc)
	ModuleCdc.Seal()
}
