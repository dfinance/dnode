package keeper

import (
	"encoding/json"

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
	for _, issue := range state.Issues {
		k.storeIssue(ctx, issue.ID, issue.Issue)
	}

	// withdraws
	for _, withdraw := range state.Withdraws {
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
	for _, issue := range k.getGenesisIssues(ctx) {
		state.Issues = append(state.Issues, issue)
	}

	// withdraws
	for _, withdraw := range k.getWithdraws(ctx) {
		state.Withdraws = append(state.Withdraws, withdraw)
	}

	return k.cdc.MustMarshalJSON(state)
}
