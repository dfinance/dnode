package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// RegisterCodec registers concrete types on the Amino codec.
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgPostPrice{}, "oracle/MsgPostPrice", nil)
	cdc.RegisterConcrete(MsgAddOracle{}, "oracle/MsgAddOracle", nil)
	cdc.RegisterConcrete(MsgSetOracles{}, "oracle/MsgSetOracles", nil)
	cdc.RegisterConcrete(MsgAddAsset{}, "oracle/MsgAddAsset", nil)
	cdc.RegisterConcrete(MsgSetAsset{}, "oracle/MsgSetAsset", nil)
}

// generic sealed codec to be used throughout module
var ModuleCdc *codec.Codec

func init() {
	cdc := codec.New()
	RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	ModuleCdc = cdc.Seal()
}
