package types

import (
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/dfinance/dnode/x/core/msmodule"
)

var ModuleCdc *codec.Codec

// RegisterCodec registers module specific messages.
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgSubmitCall{}, ModuleName+"/SubmitCall", nil)
	cdc.RegisterConcrete(MsgConfirmCall{}, ModuleName+"/ConfirmCall", nil)
	cdc.RegisterConcrete(MsgRevokeConfirm{}, ModuleName+"/RevokeConfirm", nil)

	cdc.RegisterInterface((*msmodule.MsMsg)(nil), nil)
}

func init() {
	cdc := codec.New()
	RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	ModuleCdc = cdc.Seal()
}
