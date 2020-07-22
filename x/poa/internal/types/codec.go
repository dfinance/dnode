package types

import (
	"github.com/cosmos/cosmos-sdk/codec"

	//msExport "github.com/dfinance/dnode/x/multisig/export"

	msClient "github.com/dfinance/dnode/x/multisig/client"
)

const (
	CodecNameMsgAddValidator     = ModuleName + "/AddValidator"
	CodecNameMsgRemoveValidator  = ModuleName + "/RemoveValidator"
	CodecNameMsgReplaceValidator = ModuleName + "/ReplaceValidator"
)

var ModuleCdc *codec.Codec

// RegisterCodec registers module specific messages.
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgAddValidator{}, CodecNameMsgAddValidator, nil)
	cdc.RegisterConcrete(MsgRemoveValidator{}, CodecNameMsgRemoveValidator, nil)
	cdc.RegisterConcrete(MsgReplaceValidator{}, CodecNameMsgReplaceValidator, nil)
}

func init() {
	cdc := codec.New()
	RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	ModuleCdc = cdc.Seal()

	msClient.RegisterMultiSigTypeCodec(MsgAddValidator{}, CodecNameMsgAddValidator)
	msClient.RegisterMultiSigTypeCodec(MsgRemoveValidator{}, CodecNameMsgRemoveValidator)
	msClient.RegisterMultiSigTypeCodec(MsgReplaceValidator{}, CodecNameMsgReplaceValidator)
}
