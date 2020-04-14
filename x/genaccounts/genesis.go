package genaccounts

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authExported "github.com/cosmos/cosmos-sdk/x/auth/exported"

	"github.com/dfinance/dnode/x/genaccounts/internal/types"
)

// InitGenesis initializes accounts and deliver genesis transactions.
func InitGenesis(ctx sdk.Context, _ *codec.Codec, accountKeeper types.AccountKeeper, genesisState GenesisState) {
	genesisState.Sanitize()

	for _, acc := range genesisState {
		newAcc := accountKeeper.NewAccount(ctx, &acc)
		accountKeeper.SetAccount(ctx, newAcc)
	}
}

// ExportGenesis exports genesis for all accounts.
func ExportGenesis(ctx sdk.Context, accountKeeper types.AccountKeeper) GenesisState {
	genesisState := types.GenesisState{}

	accountKeeper.IterateAccounts(ctx,
		func(acc authExported.Account) (stop bool) {
			bAcc := auth.NewBaseAccount(acc.GetAddress(), acc.GetCoins(), acc.GetPubKey(), acc.GetAccountNumber(), acc.GetSequence())
			genesisState = append(genesisState, *bAcc)
			return false
		},
	)

	return genesisState
}
