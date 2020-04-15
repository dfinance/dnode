package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authExported "github.com/cosmos/cosmos-sdk/x/auth/exported"
)

// AccountKeeper defines the expected account keeper (noalias)
type AccountKeeper interface {
	NewAccount(sdk.Context, authExported.Account) authExported.Account
	SetAccount(sdk.Context, authExported.Account)
	IterateAccounts(ctx sdk.Context, process func(authExported.Account) (stop bool))
}
