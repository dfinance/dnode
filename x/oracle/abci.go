package oracle

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

// EndBlocker processes active, rejected calls and their confirmations.
func EndBlocker(ctx sdk.Context, k Keeper) []abci.ValidatorUpdate {
	// TODO val_state_change.go is relevant if we want to rotate the oracle set

	// Running in the end blocker ensures that prices will update at most once per block,
	// which seems preferable to having state storage values change in response to multiple transactions
	// which occur during a block
	//TODO use an iterator and update the prices for all assets in the store
	if err := k.SetCurrentPrices(ctx); err != nil {
		panic(err.Error())
	}

	return []abci.ValidatorUpdate{}
}
