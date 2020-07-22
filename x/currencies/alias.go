package currencies

import (
	"github.com/dfinance/dnode/x/currencies/internal/keeper"
	"github.com/dfinance/dnode/x/currencies/internal/types"
)

type (
	Keeper              = keeper.Keeper
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
	ModuleName   = types.ModuleName
	StoreKey     = types.StoreKey
	RouterKey    = types.RouterKey
	GovRouterKey = types.GovRouterKey
	//
	QueryWithdraws = types.QueryWithdraws
	QueryWithdraw  = types.QueryWithdraw
	QueryIssue     = types.QueryIssue
	QueryCurrency  = types.QueryCurrency
	// Event types, attribute types and values
	EventTypesIssue    = types.EventTypesIssue
	EventTypesWithdraw = types.EventTypesWithdraw
	//
	AttributeDenom      = types.AttributeDenom
	AttributeAmount     = types.AttributeAmount
	AttributeIssueId    = types.AttributeIssueId
	AttributeWithdrawId = types.AttributeWithdrawId
	AttributeSender     = types.AttributeSender
)

var (
	// variable aliases
	ModuleCdc            = types.ModuleCdc
	AvailablePermissions = types.AvailablePermissions
	// function aliases
	RegisterCodec          = types.RegisterCodec
	NewKeeper              = keeper.NewKeeper
	NewQuerier             = keeper.NewQuerier
	RegisterInvariants     = keeper.RegisterInvariants
	NewMsgIssueCurrency    = types.NewMsgIssueCurrency
	NewMsgWithdrawCurrency = types.NewMsgWithdrawCurrency
	NewAddCurrencyProposal = types.NewAddCurrencyProposal
	// perms requests
	RequestCCStoragePerms = types.RequestCCStoragePerms
	// errors
	ErrInternal            = types.ErrInternal
	ErrWrongDenom          = types.ErrWrongDenom
	ErrWrongAmount         = types.ErrWrongAmount
	ErrWrongIssueID        = types.ErrWrongIssueID
	ErrWrongWithdrawID     = types.ErrWrongWithdrawID
	ErrWrongPegZoneSpender = types.ErrWrongPegZonePayee
	ErrGovInvalidProposal  = types.ErrGovInvalidProposal

	// Mint denom and event type when mint happen.
	MintDenom = types.MintDenom
	MintEvent = types.MintEventType
)
