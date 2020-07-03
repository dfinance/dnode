package types

import (
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/dfinance/dnode/x/core/msmodule"
)

var ModuleCdc *codec.Codec

// RegisterCodec registers module specific messages.
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgSubmitCall{}, "multisig/submit-call", nil)
	cdc.RegisterConcrete(MsgConfirmCall{}, "multisig/confirm-call", nil)
	cdc.RegisterConcrete(MsgRevokeConfirm{}, "multisig/revoke-confirm", nil)

	cdc.RegisterInterface((*msmodule.MsMsg)(nil), nil)
}

func init() {
	cdc := codec.New()
	RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	ModuleCdc = cdc.Seal()
}
