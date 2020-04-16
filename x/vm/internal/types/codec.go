package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// Register concrete types on codec codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgDeployModule{}, ModuleName+"/MsgDeployModule", nil)
	cdc.RegisterConcrete(MsgExecuteScript{}, ModuleName+"/MsgExecuteScript", nil)
	cdc.RegisterConcrete(TestProposal{}, ModuleName+"/TestProposal", nil)
}

// module codec
var ModuleCdc *codec.Codec

func init() {
	ModuleCdc = codec.New()
	RegisterCodec(ModuleCdc)
	ModuleCdc.Seal()
}
