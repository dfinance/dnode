package app

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/supply"
	"github.com/dfinance/dnode/cmd/config/genesis/defaults"
	"github.com/dfinance/dnode/x/ccstorage"
	"github.com/dfinance/dnode/x/vmauth"
)

// mainnetExportAddZeroHeightOptions updates options map per module for Mainnet v1.0 migration.
// Options removes all XFI tokens, renames SXFI -> XFI and withdraw all current rewards.
func mainnetExportAddZeroHeightOptions(optsMap map[string]interface{}) (map[string]interface{}, error) {
	const (
		oldStakingDenom = "sxfi"
		newStakingDenom = "xfi"
	)
	var (
		denomsToRemove = []string{"xfi", "usdt", "btc"}
	)

	// Supply
	{
		moduleName := supply.ModuleName
		optsObj, found := optsMap[moduleName]
		if !found {
			return nil, fmt.Errorf("module %s: options not found", moduleName)
		}
		opts, ok := optsObj.(supply.SquashOptions)
		if !ok {
			return nil, fmt.Errorf("module %s: options type assert failed: %T", moduleName, optsObj)
		}

		for _, denom := range denomsToRemove {
			if err := opts.SetDenomOp(denom, true, "", "0"); err != nil {
				return nil, fmt.Errorf("module %s: %w", moduleName, err)
			}
		}
		if err := opts.SetDenomOp(oldStakingDenom, false, newStakingDenom, "0"); err != nil {
			return nil, fmt.Errorf("module %s: %w", moduleName, err)
		}
		optsMap[moduleName] = opts
	}
	// VMAuth
	{
		moduleName := vmauth.ModuleName
		optsObj, found := optsMap[moduleName]
		if !found {
			return nil, fmt.Errorf("module %s: options not found", moduleName)
		}
		opts, ok := optsObj.(vmauth.SquashOptions)
		if !ok {
			return nil, fmt.Errorf("module %s: options type assert failed: %T", moduleName, optsObj)
		}

		for _, denom := range denomsToRemove {
			if err := opts.SetAccountBalanceOp(denom, true, ""); err != nil {
				return nil, fmt.Errorf("module %s: %w", moduleName, err)
			}
		}
		if err := opts.SetAccountBalanceOp(oldStakingDenom, false, newStakingDenom); err != nil {
			return nil, fmt.Errorf("module %s: %w", moduleName, err)
		}
		optsMap[moduleName] = opts
	}
	// Staking
	{
		moduleName := staking.ModuleName
		optsObj, found := optsMap[moduleName]
		if !found {
			return nil, fmt.Errorf("module %s: options not found", moduleName)
		}
		opts, ok := optsObj.(staking.SquashOptions)
		if !ok {
			return nil, fmt.Errorf("module %s: options type assert failed: %T", moduleName, optsObj)
		}

		if err := opts.SetParamsOp(newStakingDenom); err != nil {
			return nil, fmt.Errorf("module %s: %w", moduleName, err)
		}
		optsMap[moduleName] = opts
	}
	// Distribution
	{
		moduleName := distribution.ModuleName
		optsObj, found := optsMap[moduleName]
		if !found {
			return nil, fmt.Errorf("module %s: options not found", moduleName)
		}
		opts, ok := optsObj.(distribution.SquashOptions)
		if !ok {
			return nil, fmt.Errorf("module %s: options type assert failed: %T", moduleName, optsObj)
		}

		if err := opts.SetDecCoinOp(newStakingDenom, true, ""); err != nil {
			return nil, fmt.Errorf("module %s: %w", moduleName, err)
		}
		if err := opts.SetDecCoinOp(oldStakingDenom, false, newStakingDenom); err != nil {
			return nil, fmt.Errorf("module %s: %w", moduleName, err)
		}
		if err := opts.SetRewardOps(true, true); err != nil {
			return nil, fmt.Errorf("module %s: %w", moduleName, err)
		}
		optsMap[moduleName] = opts
	}
	// Mint
	{
		moduleName := mint.ModuleName
		optsObj, found := optsMap[moduleName]
		if !found {
			return nil, fmt.Errorf("module %s: options not found", moduleName)
		}
		opts, ok := optsObj.(mint.SquashOptions)
		if !ok {
			return nil, fmt.Errorf("module %s: options type assert failed: %T", moduleName, optsObj)
		}

		if err := opts.SetParamsOp(newStakingDenom); err != nil {
			return nil, fmt.Errorf("module %s: %w", moduleName, err)
		}
		optsMap[moduleName] = opts
	}
	// Gov
	{
		moduleName := gov.ModuleName
		optsObj, found := optsMap[moduleName]
		if !found {
			return nil, fmt.Errorf("module %s: options not found", moduleName)
		}
		opts, ok := optsObj.(gov.SquashOptions)
		if !ok {
			return nil, fmt.Errorf("module %s: options type assert failed: %T", moduleName, optsObj)
		}

		if err := opts.SetParamsOp(defaults.GovMinDepositAmount + newStakingDenom); err != nil {
			return nil, fmt.Errorf("module %s: %w", moduleName, err)
		}
		optsMap[moduleName] = opts
	}
	// CCStorage
	{
		moduleName := ccstorage.ModuleName
		optsObj, found := optsMap[moduleName]
		if !found {
			return nil, fmt.Errorf("module %s: options not found", moduleName)
		}
		opts, ok := optsObj.(ccstorage.SquashOptions)
		if !ok {
			return nil, fmt.Errorf("module %s: options type assert failed: %T", moduleName, optsObj)
		}

		if err := opts.SetSupplyOperation(true); err != nil {
			return nil, fmt.Errorf("module %s: %w", moduleName, err)
		}
		optsMap[moduleName] = opts
	}

	return optsMap, nil
}
