package types

import "github.com/cosmos/cosmos-sdk/codec"

var ModuleCdc = codec.New()

// RegisterCodec registers module specific messages.
func RegisterCodec(cdc *codec.Codec) {}
