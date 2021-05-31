package app

import (
	"fmt"
	"os"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/mint"
	stakingTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func mainnetExportRemoveAllValidators(ctx sdk.Context, app *DnServiceApp) error {
	fakeTime := ctx.BlockTime().Add(stakingTypes.MaxUnbondingTime)

	app.Logger().Info("Removing all validators with delegations refund:")

	app.Logger().Info("  Stopping all active redelegations:")
	reds := make([]stakingTypes.Redelegation, 0)
	app.stakingKeeper.IterateRedelegations(ctx, func(_ int64, red stakingTypes.Redelegation) (stop bool) {
		reds = append(reds, red)
		return false
	})
	for redIdx, red := range reds {
		app.Logger().Info(fmt.Sprintf("    RED [%d]: %s, %s -> %s", redIdx, red.DelegatorAddress, red.ValidatorSrcAddress, red.ValidatorDstAddress))
		if _, err := app.stakingKeeper.CompleteRedelegationWithAmount(ctx, red.DelegatorAddress, red.ValidatorSrcAddress, red.ValidatorDstAddress, fakeTime); err != nil {
			return fmt.Errorf("stopping redelegation (%s, %s -> %s): %w",
				red.DelegatorAddress, red.ValidatorSrcAddress, red.ValidatorDstAddress, err,
			)
		}
	}

	app.Logger().Info("  Removing all validators:")
	for valIdx, val := range app.stakingKeeper.GetAllValidators(ctx) {
		app.Logger().Info(fmt.Sprintf("    Validator [%d]: %s (%s):", valIdx, val.GetMoniker(), val.GetOperator()))
		if err := app.stakingKeeper.ForceUnbondValidator(ctx, val); err != nil {
			return fmt.Errorf("ForceUnbondValidator of %s (%s): %w", val.GetMoniker(), val.GetOperator(), err)
		}
	}

	app.Logger().Info("  Stopping all active undelegations:")
	ubds := make([]stakingTypes.UnbondingDelegation, 0)
	app.stakingKeeper.IterateUnbondingDelegations(ctx, func(_ int64, ubd stakingTypes.UnbondingDelegation) (stop bool) {
		ubds = append(ubds, ubd)
		return false
	})
	for ubdIdx, ubd := range ubds {
		app.Logger().Info(fmt.Sprintf("    UBD [%d]: %s for %s", ubdIdx, ubd.DelegatorAddress, ubd.ValidatorAddress))
		if _, err := app.stakingKeeper.CompleteUnbondingWithAmount(ctx, ubd.DelegatorAddress, ubd.ValidatorAddress, fakeTime); err != nil {
			return fmt.Errorf("stopping undelegation (%s, %s): %w",
				ubd.DelegatorAddress, ubd.ValidatorAddress, err,
			)
		}
	}

	app.Logger().Info("  Applying validators set updates")
	_ = app.stakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)

	app.Logger().Info("  Checking invariants")
	if err := app.checkInvariants(ctx); err != nil {
		return fmt.Errorf("invariants check: %w", err)
	}

	return nil
}

// mainnetExportProcessBalances builds MainnetBalanceReport, mints and transfers negative diffs.
func mainnetExportProcessBalances(ctx sdk.Context, app *DnServiceApp) error {
	const (
		issueDenom   = "sxfi"
		bondingDenom = "xfi"
		lpDenom      = "lpt"
	)

	app.Logger().Info("Initial vs current balance processing:")

	stakerReportPath := os.Getenv("DN_ZHP_STAKERREPORT_PATH")
	reportOutputPrefix := os.Getenv("DN_ZHP_REPORTOUTPUT_PREFIX")
	if stakerReportPath == "" {
		app.Logger().Info("  WARN: Staker report path not provided, skipping it (DN_ZHP_STAKERREPORT_PATH)")
	}
	if reportOutputPrefix == "" {
		return fmt.Errorf("envVar %q: not set", "DN_ZHP_REPORTOUTPUT_PREFIX")
	}

	// build report
	app.Logger().Info("  Report: build and validation")
	report, err := NewMainnetBalanceReport(ctx, app, issueDenom, lpDenom, bondingDenom, lpDenom, stakerReportPath)
	if err != nil {
		return fmt.Errorf("NewMainnetBalanceReport: %w", err)
	}
	if err := report.Verify(); err != nil {
		return fmt.Errorf("report validation: %w", err)
	}

	// build results
	app.Logger().Info("  Results: build and save")
	results := report.GetResults()
	if err := results.SaveToCSV(reportOutputPrefix + "data.csv"); err != nil {
		return fmt.Errorf("report results CSV export: %w", err)
	}

	// build stats
	app.Logger().Info("  Stats: build and save")
	stats := NewMainnetBalanceStats(ctx, app, results)
	if err := stats.SaveToJSON(reportOutputPrefix + "stats.json"); err != nil {
		return fmt.Errorf("report results stats JSON export: %w", err)
	}

	// mint
	app.Logger().Info("  Mint: mint and distribute compensations")
	if err := app.mintKeeper.MintCoins(ctx, stats.GetTotalMintCoins()); err != nil {
		return fmt.Errorf("minting bonding coins: %w", err)
	}

	bondingSupply := sdk.NewCoin(bondingDenom, app.supplyKeeper.GetSupply(ctx).GetTotal().AmountOf(bondingDenom))
	if err := app.ccsKeeper.IncreaseCurrencySupply(ctx, bondingSupply); err != nil {
		return fmt.Errorf("increasing ccStorage supply: %w", err)
	}

	for _, acc := range stats.AccBondingMints {
		if err := app.supplyKeeper.SendCoinsFromModuleToAccount(ctx, mint.ModuleName, acc.AccAddress, sdk.NewCoins(acc.MintCoin)); err != nil {
			return fmt.Errorf("sending minted coins (%s) to (%s): %w", acc.MintCoin, acc.AccAddress, err)
		}
	}

	// check the invariants
	app.Logger().Info("  Checking invariants")
	if err := app.checkInvariants(ctx); err != nil {
		return fmt.Errorf("post invariants check: %w", err)
	}

	return nil
}
