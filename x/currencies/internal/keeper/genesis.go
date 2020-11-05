package keeper

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dfinance/dnode/x/currencies/internal/types"
)

// InitGenesis inits module genesis state: creates currencies.
func (k Keeper) InitGenesis(ctx sdk.Context, data json.RawMessage) {
	k.modulePerms.AutoCheck(types.PermInit)

	state := types.GenesisState{}
	k.cdc.MustUnmarshalJSON(data, &state)

	// validate again knowing current blockHeight
	if err := state.Validate(ctx.BlockTime()); err != nil {
		panic(err)
	}

	// last withdrawID
	if state.LastWithdrawID != nil {
		k.setLastWithdrawID(ctx, *state.LastWithdrawID)
	}

	// issues
	for i, issue := range state.Issues {
		if !k.ccsKeeper.HasCurrency(ctx, issue.Coin.Denom) {
			panic(fmt.Errorf("issue[%d] denom %q: currency not found", i, issue.Coin.Denom))
		}

		k.storeIssue(ctx, issue.ID, issue.Issue)
	}

	// withdraws
	for i, withdraw := range state.Withdraws {
		if !k.ccsKeeper.HasCurrency(ctx, withdraw.Coin.Denom) {
			panic(fmt.Errorf("withdraw[%d] denom %q: currency not found", i, withdraw.Coin.Denom))
		}

		k.storeWithdraw(ctx, withdraw)
	}
}

// ExportGenesis exports module genesis state using current params state.
func (k Keeper) ExportGenesis(ctx sdk.Context) json.RawMessage {
	k.modulePerms.AutoCheck(types.PermRead)

	state := types.GenesisState{
		Issues:    make([]types.GenesisIssue, 0),
		Withdraws: types.Withdraws{},
	}

	// last withdrawID
	if k.hasLastWithdrawID(ctx) {
		lastID := k.getLastWithdrawID(ctx)
		state.LastWithdrawID = &lastID
	}

	// issues
	state.Issues = append(state.Issues, k.GetGenesisIssues(ctx)...)

	// withdraws
	state.Withdraws = append(state.Withdraws, k.getWithdraws(ctx)...)

	return k.cdc.MustMarshalJSON(state)
}
