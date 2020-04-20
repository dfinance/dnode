package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// RegisterCodec registers concrete types on the Amino codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgPostPrice{}, "pricefeed/MsgPostPrice", nil)
	cdc.RegisterConcrete(MsgAddPriceFeed{}, "pricefeed/MsgAddPriceFeed", nil)
	cdc.RegisterConcrete(MsgSetPriceFeeds{}, "pricefeed/MsgSetPriceFeeds", nil)
	cdc.RegisterConcrete(MsgAddAsset{}, "pricefeed/MsgAddAsset", nil)
	cdc.RegisterConcrete(MsgSetAsset{}, "pricefeed/MsgSetAsset", nil)
}

// generic sealed codec to be used throughout module
var ModuleCdc *codec.Codec

func init() {
	cdc := codec.New()
	RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	ModuleCdc = cdc.Seal()
}
