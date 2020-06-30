package types

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/gov"
)

var ModuleCdc *codec.Codec

// RegisterCodec registers module specific messages.
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgIssueCurrency{}, fmt.Sprintf("%s/issue-currency", ModuleName), nil)
	cdc.RegisterConcrete(MsgWithdrawCurrency{}, fmt.Sprintf("%s/withdraw-currency", ModuleName), nil)
	cdc.RegisterConcrete(AddCurrencyProposal{}, fmt.Sprintf("%s/AddCurrencyProposal", ModuleName), nil)
}

func init() {
	cdc := codec.New()
	RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	ModuleCdc = cdc.Seal()

	gov.RegisterProposalType(ProposalTypeAddCurrency)
	gov.RegisterProposalTypeCodec(AddCurrencyProposal{}, GovRouterKey+"/AddCurrencyProposal")
}
