package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/gov"
)

// Register concrete types on codec codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgDeployModule{}, ModuleName+"/MsgDeployModule", nil)
	cdc.RegisterConcrete(MsgExecuteScript{}, ModuleName+"/MsgExecuteScript", nil)

	cdc.RegisterInterface((*PlannedProposal)(nil), nil)
	cdc.RegisterConcrete(TestProposal{}, ModuleName+"/TestProposal", nil)
	cdc.RegisterConcrete(StdlibUpdateProposal{}, ModuleName+"/StdlibUpdateProposal", nil)
}

// module codec
var ModuleCdc *codec.Codec

func init() {
	ModuleCdc = codec.New()
	RegisterCodec(ModuleCdc)
	ModuleCdc.Seal()

	gov.RegisterProposalType(ProposalTypeStdlibUpdate)
	gov.RegisterProposalTypeCodec(StdlibUpdateProposal{}, ModuleName+"/StdlibUpdateProposal")
}
