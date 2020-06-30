package currencies

import (
	"github.com/dfinance/dnode/x/currencies/internal/keeper"
	"github.com/dfinance/dnode/x/currencies/internal/types"
)

type (
	Keeper              = keeper.Keeper
	Currency            = types.Currency
	Issue               = types.Issue
	Withdraw            = types.Withdraw
	Withdraws           = types.Withdraws
	MsgIssueCurrency    = types.MsgIssueCurrency
	MsgWithdrawCurrency = types.MsgWithdrawCurrency
	CurrencyReq         = types.CurrencyReq
	IssueReq            = types.IssueReq
	WithdrawsReq        = types.WithdrawsReq
	WithdrawReq         = types.WithdrawReq
)

const (
	ModuleName = types.ModuleName
	RouterKey  = types.RouterKey
	StoreKey   = types.StoreKey
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
	NewMsgIssueCurrency    = types.NewMsgIssueCurrency
	NewMsgWithdrawCurrency = types.NewMsgWithdrawCurrency
	// errors
	ErrInternal            = types.ErrInternal
	ErrWrongDenom          = types.ErrWrongDenom
	ErrWrongAmount         = types.ErrWrongAmount
	ErrWrongIssueID        = types.ErrWrongIssueID
	ErrWrongWithdrawID     = types.ErrWrongWithdrawID
	ErrWrongPegZoneSpender = types.ErrWrongPegZoneSpender
	ErrIncorrectDecimals   = types.ErrIncorrectDecimals
)
