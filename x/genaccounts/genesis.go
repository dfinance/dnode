package genaccounts

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authExported "github.com/cosmos/cosmos-sdk/x/auth/exported"

	"github.com/dfinance/dnode/x/genaccounts/internal/types"
)

// InitGenesis initializes accounts and deliver genesis transactions.
func InitGenesis(ctx sdk.Context, _ *codec.Codec, accountKeeper types.AccountKeeper, genesisState GenesisState) {
	genesisState.Sanitize()

	for _, ga := range genesisState {
		acc := ga.ToAccount()
		acc = accountKeeper.NewAccount(ctx, acc)
		accountKeeper.SetAccount(ctx, acc)
	}
}

// ExportGenesis exports genesis for all accounts.
func ExportGenesis(ctx sdk.Context, accountKeeper types.AccountKeeper) GenesisState {
	gAccounts := GenesisAccounts{}

	accountKeeper.IterateAccounts(ctx,
		func(acc authExported.Account) (stop bool) {
			gAccount, err := NewGenesisAccountI(acc)
			if err != nil {
				panic(err)
			}
			gAccounts = append(gAccounts, gAccount)
			return false
		},
	)

	return GenesisState(gAccounts)
}
