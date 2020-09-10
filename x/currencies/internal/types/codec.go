package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/gov"

	msClient "github.com/dfinance/dnode/x/multisig/client"
)

const (
	CodecNameMsgIssueCurrency    = ModuleName + "/IssueCurrency"
	CodecNameMsgWithdrawCurrency = ModuleName + "/WithdrawCurrency"
	CodecNameAddCurrencyProposal = ModuleName + "/AddCurrencyProposal"
	CodecNameMsgUnstakeCurrency  = ModuleName + "/UnstakeCurrency"
)

var ModuleCdc *codec.Codec

// RegisterCodec registers module specific messages.
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterConcrete(MsgIssueCurrency{}, CodecNameMsgIssueCurrency, nil)
	cdc.RegisterConcrete(MsgWithdrawCurrency{}, CodecNameMsgWithdrawCurrency, nil)
	cdc.RegisterConcrete(AddCurrencyProposal{}, CodecNameAddCurrencyProposal, nil)
	cdc.RegisterConcrete(MsgUnstakeCurrency{}, CodecNameMsgUnstakeCurrency, nil)
}

func init() {
	cdc := codec.New()
	RegisterCodec(cdc)
	codec.RegisterCrypto(cdc)
	ModuleCdc = cdc.Seal()

	msClient.RegisterMultiSigTypeCodec(MsgIssueCurrency{}, CodecNameMsgIssueCurrency)
	msClient.RegisterMultiSigTypeCodec(MsgUnstakeCurrency{}, CodecNameMsgUnstakeCurrency)
	msClient.RegisterMultiSigTypeCodec(MsgWithdrawCurrency{}, CodecNameMsgWithdrawCurrency)

	gov.RegisterProposalType(ProposalTypeAddCurrency)
	gov.RegisterProposalTypeCodec(AddCurrencyProposal{}, CodecNameAddCurrencyProposal)
}
