package types

import "github.com/cosmos/cosmos-sdk/codec"

var ModuleCdc = codec.New()

func RegisterCodec(cdc *codec.Codec) {}
