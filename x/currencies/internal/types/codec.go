package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/gov"
)

var ModuleCdc *codec.Codec

// RegisterCodec registers module specific messages.
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgIssueCurrency{}, ModuleName+"/IssueCurrency", nil)
	cdc.RegisterConcrete(MsgWithdrawCurrency{}, ModuleName+"/WithdrawCurrency", nil)
	cdc.RegisterConcrete(AddCurrencyProposal{}, ModuleName+"/AddCurrencyProposal", nil)
}

func init() {
	cdc := codec.New()
	RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	ModuleCdc = cdc.Seal()

	gov.RegisterProposalType(ProposalTypeAddCurrency)
	gov.RegisterProposalTypeCodec(AddCurrencyProposal{}, GovRouterKey+"/AddCurrencyProposal")
}
