package currencies

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"wings-blockchain/x/currencies/msgs"
)

func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(msgs.MsgIssueCurrency{},   "currencies/issue-currency", nil)
	cdc.RegisterConcrete(msgs.MsgDestroyCurrency{}, "currencies/destroy-currency", nil)
}
