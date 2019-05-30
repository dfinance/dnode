package poa

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"wings-blockchain/x/poa/msgs"
)

func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(msgs.MsgAddValidator{}, 	 msgs.MsgAddValidatorType,	  nil)
	cdc.RegisterConcrete(msgs.MsgRemoveValidator{},  msgs.MsgRemoveValidatorType, nil)
	cdc.RegisterConcrete(msgs.MsgReplaceValidator{}, msgs.MsgReplaceValidatorType,nil)
}
