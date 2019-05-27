package currencies

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"wings-blockchain/x/currencies/msgs"
)

func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(msgs.MsgIssueCurrency{},   "currencies/IssueCurrency", nil)
	cdc.RegisterConcrete(msgs.MsgDestroyCurrency{}, "currencies/DestroyCurrency", nil)
}
