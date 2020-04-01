package clitester

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func makeCodec() *codec.Codec {
	var cdc = codec.New()
	ModuleBasics.RegisterCodec(cdc) // register all module codecs.
	sdk.RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)

	return cdc
}

func trimCliOutput(output []byte) []byte {
	for i := 0; i < len(output); i++ {
		if output[i] == '{' {
			output = output[i:]
			break
		}
	}

	return output
}
