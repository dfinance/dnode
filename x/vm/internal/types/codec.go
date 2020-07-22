package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/gov"
)

var ModuleCdc *codec.Codec

// RegisterCodec registers module specific messages.
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgDeployModule{}, ModuleName+"/MsgDeployModule", nil)
	cdc.RegisterConcrete(MsgExecuteScript{}, ModuleName+"/MsgExecuteScript", nil)

	cdc.RegisterInterface((*PlannedProposal)(nil), nil)
	cdc.RegisterConcrete(TestProposal{}, ModuleName+"/TestProposal", nil)
	cdc.RegisterConcrete(StdlibUpdateProposal{}, ModuleName+"/StdlibUpdateProposal", nil)
}

func init() {
	ModuleCdc = codec.New()
	RegisterCodec(ModuleCdc)
	ModuleCdc.Seal()

	gov.RegisterProposalType(ProposalTypeStdlibUpdate)
	gov.RegisterProposalTypeCodec(StdlibUpdateProposal{}, GovRouterKey+"/StdlibUpdateProposal")
}
