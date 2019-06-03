package multisig

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"wings-blockchain/x/multisig/msgs"
)

func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(msgs.MsgSubmitCall{}, "multisig/submit-call",nil)
	cdc.RegisterConcrete(msgs.MsgConfirmCall{}, "multisig/confirm-call",nil)
}
