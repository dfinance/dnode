package types

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
)

var ModuleCdc = codec.New()

func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgCreateMarket{}, fmt.Sprintf("%s/%s", ModuleName, MsgCreateMarket{}.Type()), nil)
}
