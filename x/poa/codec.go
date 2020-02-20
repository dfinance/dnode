package poa

import (
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/WingsDao/wings-blockchain/x/poa/msgs"
)

// Registering amino types for PoA messages
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(msgs.MsgAddValidator{}, msgs.MsgAddValidatorType, nil)
	cdc.RegisterConcrete(msgs.MsgRemoveValidator{}, msgs.MsgRemoveValidatorType, nil)
	cdc.RegisterConcrete(msgs.MsgReplaceValidator{}, msgs.MsgReplaceValidatorType, nil)
}

var ModuleCdc *codec.Codec

// Initialize codec before everything else.
func init() {
	ModuleCdc = codec.New()
	RegisterCodec(ModuleCdc)
	codec.RegisterCrypto(ModuleCdc)
	ModuleCdc.Seal()
}
