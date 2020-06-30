package currencies

import (
	"github.com/dfinance/dnode/x/currencies/internal/keeper"
	"github.com/dfinance/dnode/x/currencies/internal/types"
)

type (
	Keeper              = keeper.Keeper
	GenesisState        = types.GenesisState
	Currency            = types.Currency
	Issue               = types.Issue
	Withdraw            = types.Withdraw
	Withdraws           = types.Withdraws
	MsgIssueCurrency    = types.MsgIssueCurrency
	MsgWithdrawCurrency = types.MsgWithdrawCurrency
	AddCurrencyProposal = types.AddCurrencyProposal
	CurrencyReq         = types.CurrencyReq
	IssueReq            = types.IssueReq
	WithdrawsReq        = types.WithdrawsReq
	WithdrawReq         = types.WithdrawReq
)

const (
	ModuleName        = types.ModuleName
	StoreKey          = types.StoreKey
	DefaultParamspace = types.DefaultParamspace
	RouterKey         = types.RouterKey
	GovRouterKey      = types.GovRouterKey
	//
	QueryWithdraws = types.QueryWithdraws
	QueryWithdraw  = types.QueryWithdraw
	QueryIssue     = types.QueryIssue
	QueryCurrency  = types.QueryCurrency
)

var (
	// variable aliases
	ModuleCdc = types.ModuleCdc
	// function aliases
	RegisterCodec          = types.RegisterCodec
	NewKeeper              = keeper.NewKeeper
	NewQuerier             = keeper.NewQuerier
	DefaultGenesisState    = types.DefaultGenesisState
	NewMsgIssueCurrency    = types.NewMsgIssueCurrency
	NewMsgWithdrawCurrency = types.NewMsgWithdrawCurrency
	NewAddCurrencyProposal = types.NewAddCurrencyProposal
	// errors
	ErrInternal            = types.ErrInternal
	ErrWrongDenom          = types.ErrWrongDenom
	ErrWrongAmount         = types.ErrWrongAmount
	ErrWrongIssueID        = types.ErrWrongIssueID
	ErrWrongWithdrawID     = types.ErrWrongWithdrawID
	ErrWrongPegZoneSpender = types.ErrWrongPegZoneSpender
	ErrIncorrectDecimals   = types.ErrIncorrectDecimals
	ErrWrongParams         = types.ErrWrongParams
	ErrGovInvalidProposal  = types.ErrGovInvalidProposal
)
