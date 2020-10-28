package app

import (
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/staking"
	abci "github.com/tendermint/tendermint/abci/types"
	tmTypes "github.com/tendermint/tendermint/types"
)

// Exports genesis and validators.
func (app *DnServiceApp) ExportAppStateAndValidators(forZeroHeight bool, jailWhiteList []string,
) (appState json.RawMessage, validators []tmTypes.GenesisValidator, retErr error) {

	var err error

	// create a new context
	ctx := app.NewContext(true, abci.Header{Height: app.LastBlockHeight()})

	// zero-height squash
	if forZeroHeight {
		if err := app.prepareForZeroHeightGenesis(ctx); err != nil {
			retErr = fmt.Errorf("preparing for zero-height: %w", err)
			return
		}
	}

	// genesis export
	genState := app.mm.ExportGenesis(ctx)
	appState, err = codec.MarshalJSONIndent(app.cdc, genState)
	if err != nil {
		retErr = fmt.Errorf("genState JSON marshal: %w", err)
		return
	}

	validators = staking.WriteValidators(ctx, app.stakingKeeper)

	return appState, validators, nil
}

func (app *DnServiceApp) checkInvariants(ctx sdk.Context) error {
	for _, invRoute := range app.crisisKeeper.Routes() {
		res, stop := invRoute.Invar(ctx)
		if stop {
			return fmt.Errorf("module %s (%s): %s", invRoute.ModuleName, invRoute.Route, res)
		}
	}

	return nil
}

func (app *DnServiceApp) prepareForZeroHeightGenesis(ctx sdk.Context) error {
	// Check invariants before
	if err := app.checkInvariants(ctx); err != nil {
		return fmt.Errorf("pre invariants check failed: %w", err)
	}

	// Cosmos SDK modules
	moduleName := distribution.ModuleName
	if err := app.distrKeeper.PrepareForZeroHeight(ctx); err != nil {
		return fmt.Errorf("module %s: %w", moduleName, err)
	}

	moduleName = staking.ModuleName
	if err := app.stakingKeeper.PrepareForZeroHeight(ctx); err != nil {
		return fmt.Errorf("module %s: %w", moduleName, err)
	}

	moduleName = slashing.ModuleName
	if err := app.slashingKeeper.PrepareForZeroHeight(ctx); err != nil {
		return fmt.Errorf("module %s: %w", moduleName, err)
	}

	// Check invariants after
	if err := app.checkInvariants(ctx); err != nil {
		return fmt.Errorf("post invariants check failed: %w", err)
	}

	return nil
}
