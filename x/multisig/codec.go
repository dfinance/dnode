// Registering amino types for multisignature usage.
package multisig

import (
	"github.com/cosmos/cosmos-sdk/codec"

	"wings-blockchain/x/core"
	"wings-blockchain/x/multisig/msgs"
)

// Register amino types for multisig module.
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(msgs.MsgSubmitCall{}, "multisig/submit-call", nil)
	cdc.RegisterConcrete(msgs.MsgConfirmCall{}, "multisig/confirm-call", nil)
	cdc.RegisterConcrete(msgs.MsgRevokeConfirm{}, "multisig/revoke-confirm", nil)

	cdc.RegisterInterface((*core.MsMsg)(nil), nil)
}

var ModuleCdc *codec.Codec

// Initialize codec before everything else.
func init() {
	ModuleCdc = codec.New()
	RegisterCodec(ModuleCdc)
	codec.RegisterCrypto(ModuleCdc)
	ModuleCdc.Seal()
}
