package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/gov"
)

// Register concrete types on codec codec
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(AddCurrencyProposal{}, ModuleName+"/AddCurrencyProposal", nil)
}

// module codec
var ModuleCdc *codec.Codec

func init() {
	ModuleCdc = codec.New()
	RegisterCodec(ModuleCdc)
	ModuleCdc.Seal()

	gov.RegisterProposalType(ProposalTypeAddCurrency)
	gov.RegisterProposalTypeCodec(AddCurrencyProposal{}, GovRouterKey+"/AddCurrencyProposal")
}
