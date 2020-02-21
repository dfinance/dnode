// Register codecs for currencies module.
package currencies

import (
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/WingsDao/wings-blockchain/x/currencies/msgs"
)

// Register amino types for Currencies module.
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(msgs.MsgIssueCurrency{}, "currencies/issue-currency", nil)
	cdc.RegisterConcrete(msgs.MsgDestroyCurrency{}, "currencies/destroy-currency", nil)
}

var ModuleCdc *codec.Codec

// Initialize codec before everything else.
func init() {
	ModuleCdc = codec.New()
	RegisterCodec(ModuleCdc)
	codec.RegisterCrypto(ModuleCdc)
	ModuleCdc.Seal()
}
