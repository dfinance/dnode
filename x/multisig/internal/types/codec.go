package types

import (
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/dfinance/dnode/x/core/msmodule"
)

var ModuleCdc = codec.New()

// RegisterCodec registers module specific messages.
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgSubmitCall{}, ModuleName+"/SubmitCall", nil)
	cdc.RegisterConcrete(MsgConfirmCall{}, ModuleName+"/ConfirmCall", nil)
	cdc.RegisterConcrete(MsgRevokeConfirm{}, ModuleName+"/RevokeConfirm", nil)

	cdc.RegisterInterface((*msmodule.MsMsg)(nil), nil)
}

// RegisterMultiSigTypeCodec registers an external multisig message defined in another module for the internal ModuleCdc.
func RegisterMultiSigTypeCodec(o interface{}, name string) {
	ModuleCdc.RegisterConcrete(o, name, nil)
}

func init() {
	RegisterCodec(ModuleCdc)
	codec.RegisterCrypto(ModuleCdc)
}
