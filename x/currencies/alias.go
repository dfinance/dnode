package currencies

import (
	"github.com/dfinance/dnode/x/currencies/internal/keeper"
	"github.com/dfinance/dnode/x/currencies/internal/types"
)

type (
	Keeper             = keeper.Keeper
	Currency           = types.Currency
	Issue              = types.Issue
	Destroy            = types.Destroy
	Destroys           = types.Destroys
	MsgIssueCurrency   = types.MsgIssueCurrency
	MsgDestroyCurrency = types.MsgDestroyCurrency
	IssueReq           = types.IssueReq
	DestroyReq         = types.DestroyReq
	DestroysReq        = types.DestroysReq
	CurrencyReq        = types.CurrencyReq
)

const (
	ModuleName = types.ModuleName
	RouterKey  = types.RouterKey
	StoreKey   = types.StoreKey
	//
	QueryDestroys = types.QueryDestroys
	QueryDestroy  = types.QueryDestroy
	QueryIssue    = types.QueryIssue
	QueryCurrency = types.QueryCurrency
)

var (
	// variable aliases
	ModuleCdc = types.ModuleCdc
	// function aliases
	RegisterCodec         = types.RegisterCodec
	NewKeeper             = keeper.NewKeeper
	NewQuerier            = keeper.NewQuerier
	NewMsgIssueCurrency   = types.NewMsgIssueCurrency
	NewMsgDestroyCurrency = types.NewMsgDestroyCurrency
	// errors
	ErrInternal          = types.ErrInternal
	ErrWrongDenom        = types.ErrWrongDenom
	ErrWrongAmount       = types.ErrWrongAmount
	ErrWrongIssueID      = types.ErrWrongIssueID
	ErrWrongDestroyID    = types.ErrWrongDestroyID
	ErrWrongRecipient    = types.ErrWrongRecipient
	ErrIncorrectDecimals = types.ErrIncorrectDecimals
)
