package app

import (
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/supply"
	abci "github.com/tendermint/tendermint/abci/types"
	tmTypes "github.com/tendermint/tendermint/types"

	"github.com/dfinance/dnode/x/ccstorage"
	"github.com/dfinance/dnode/x/multisig"
	"github.com/dfinance/dnode/x/orderbook"
	"github.com/dfinance/dnode/x/vmauth"
)

// Exports genesis and validators.
func (app *DnServiceApp) ExportAppStateAndValidators(forZeroHeight bool, jailWhiteList []string,
) (appState json.RawMessage, validators []tmTypes.GenesisValidator, retErr error) {

	var err error

	// create a new context
	ctx := app.NewContext(true, abci.Header{Height: app.LastBlockHeight()})

	// zero-height squash
	if forZeroHeight {
		if err := app.prepareGenesisForZeroHeight(ctx, jailWhiteList); err != nil {
			retErr = fmt.Errorf("preparing genesis for zero-height: %w", err)
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

// prepareGenesisForZeroHeight updates current context to fit zero-height genesis.
// Basically it "squashes" all height-dependent storage objects.
func (app *DnServiceApp) prepareGenesisForZeroHeight(ctx sdk.Context, jailWhiteList []string) error {
	// Check invariants before
	if err := app.checkInvariants(ctx); err != nil {
		return fmt.Errorf("pre invariants check failed: %w", err)
	}

	// Prepare PrepareForZeroHeight module functions options
	optsMap, err := prepareDefaultZeroHeightOptions(jailWhiteList)
	if err != nil {
		return fmt.Errorf("prepareDefaultZeroHeightOptions: %w", err)
	}
	//optsMap, err = setDebugZeroHeightOptions(optsMap)
	//if err != nil {
	//	return fmt.Errorf("setDebugZeroHeightOptions: %w", err)
	//}
	//optsMap, err = setMainnetZeroHeightOptionsV10(optsMap)
	//if err != nil {
	//	return fmt.Errorf("setMainnetZeroHeightOptionsV10: %w", err)
	//}

	// CCStorage
	{
		moduleName := ccstorage.ModuleName
		opts := optsMap[moduleName].(ccstorage.SquashOptions)
		if err := app.ccsKeeper.PrepareForZeroHeight(ctx, opts); err != nil {
			return fmt.Errorf("module %s: %w", moduleName, err)
		}
	}
	// Supply
	{
		moduleName := supply.ModuleName
		opts := optsMap[moduleName].(supply.SquashOptions)
		if err := app.supplyKeeper.PrepareForZeroHeight(ctx, opts); err != nil {
			return fmt.Errorf("module %s: %w", moduleName, err)
		}
	}
	// VMAuth
	{
		moduleName := vmauth.ModuleName
		opts := optsMap[moduleName].(vmauth.SquashOptions)
		if err := app.accountKeeper.PrepareForZeroHeight(ctx, opts); err != nil {
			return fmt.Errorf("module %s: %w", moduleName, err)
		}
	}
	// Staking
	{
		moduleName := staking.ModuleName
		opts := optsMap[moduleName].(staking.SquashOptions)
		if err := app.stakingKeeper.PrepareForZeroHeight(ctx, opts); err != nil {
			return fmt.Errorf("module %s: %w", moduleName, err)
		}
	}
	// Distribution
	{
		moduleName := distribution.ModuleName
		opts := optsMap[moduleName].(distribution.SquashOptions)
		if err := app.distrKeeper.PrepareForZeroHeight(ctx, opts); err != nil {
			return fmt.Errorf("module %s: %w", moduleName, err)
		}
	}
	// Slashing
	{
		moduleName := slashing.ModuleName
		if err := app.slashingKeeper.PrepareForZeroHeight(ctx); err != nil {
			return fmt.Errorf("module %s: %w", moduleName, err)
		}
	}
	// Mint
	{
		moduleName := mint.ModuleName
		opts := optsMap[moduleName].(mint.SquashOptions)
		if err := app.mintKeeper.PrepareForZeroHeight(ctx, opts); err != nil {
			return fmt.Errorf("module %s: %w", moduleName, err)
		}
	}
	// Gov
	{
		moduleName := gov.ModuleName
		opts := optsMap[moduleName].(gov.SquashOptions)
		if err := app.govKeeper.PrepareForZeroHeight(ctx, opts); err != nil {
			return fmt.Errorf("module %s: %w", moduleName, err)
		}
	}
	// MultiSig
	{
		moduleName := multisig.ModuleName
		if err := app.msKeeper.PrepareForZeroHeight(ctx); err != nil {
			return fmt.Errorf("module %s: %w", moduleName, err)
		}
	}
	// OrderBook
	{
		moduleName := orderbook.ModuleName
		if err := app.orderBookKeeper.PrepareForZeroHeight(ctx); err != nil {
			return fmt.Errorf("module %s: %w", moduleName, err)
		}
	}

	// Check invariants after
	if err := app.checkInvariants(ctx); err != nil {
		return fmt.Errorf("post invariants check failed: %w", err)
	}

	//
	if err := app.processMainnetSXFIBalance(ctx); err != nil {
		return fmt.Errorf("mainnet v1.0 processing: %w", err)
	}

	return nil
}

// prepareDefaultZeroHeightOptions returns base (default) options map per module for PrepareForZeroHeight functions.
func prepareDefaultZeroHeightOptions(jailWhiteList []string) (map[string]interface{}, error) {
	optsMap := make(map[string]interface{}, 0)

	// CCStorage
	{
		moduleName := ccstorage.ModuleName
		opts := ccstorage.NewEmptySquashOptions()
		optsMap[moduleName] = opts
	}
	// Supply
	{
		moduleName := supply.ModuleName
		opts := supply.NewEmptySquashOptions()
		optsMap[moduleName] = opts
	}
	// VMAuth
	{
		moduleName := vmauth.ModuleName
		opts := vmauth.NewEmptySquashOptions()
		optsMap[moduleName] = opts
	}
	// Staking
	{
		moduleName := staking.ModuleName
		opts := staking.NewEmptySquashOptions()
		if err := opts.SetJailWhitelistSquashOption(jailWhiteList); err != nil {
			return nil, fmt.Errorf("module %s: %w", moduleName, err)
		}
		optsMap[moduleName] = opts
	}
	// Distribution
	{
		moduleName := distribution.ModuleName
		opts := distribution.NewEmptySquashOptions()
		optsMap[moduleName] = opts
	}
	// Mint
	{
		moduleName := mint.ModuleName
		opts := mint.NewEmptySquashOptions()
		optsMap[moduleName] = opts
	}
	// Gov
	{
		moduleName := gov.ModuleName
		opts := gov.NewEmptySquashOptions()
		optsMap[moduleName] = opts
	}

	return optsMap, nil
}

// setDebugZeroHeightOptions updates options map per module for debug purposes.
// Adds a fake validator jailing all the others.
// This mod is helpful to run exported genesis locally with one up and running validator.
func setDebugZeroHeightOptions(optsMap map[string]interface{}) (map[string]interface{}, error) {
	const (
		// Values below are hardcoded according to bootstrap init_single_w_genesis.sh script values
		fakeValOperatorAddress      = "wallet17raernuazufad6q48uc5jdnqmuzsep5a03dc0n"
		fakeValMoniker              = "fakeVal"
		fakeValPubKey               = "walletvalconspub1zcjduepqu9mgrhdjfmwwalv86vdsavvvxfy8r4fmt4py8ehep252rs0acs5q93t5nm"
		fakeValSelfDelegationAmount = "1000000000000000000000000"
		//
		stakingDenom = "sxfi"
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

		if err := opts.SetDenomOp(stakingDenom, false, "", fakeValSelfDelegationAmount); err != nil {
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

		if err := opts.SetAddAccountOp(fakeValOperatorAddress, fakeValSelfDelegationAmount+stakingDenom); err != nil {
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

		accAddr, err := sdk.AccAddressFromBech32(fakeValOperatorAddress)
		if err != nil {
			return nil, fmt.Errorf("module %s: invalid fakeValOperatorAddress (%s): %w", moduleName, fakeValOperatorAddress, err)
		}
		valAddr := sdk.ValAddress(accAddr)

		if err := opts.SetAddValidatorOp(fakeValOperatorAddress, fakeValMoniker, fakeValPubKey, fakeValSelfDelegationAmount); err != nil {
			return nil, fmt.Errorf("module %s: %w", moduleName, err)
		}
		if err := opts.SetJailWhitelistSquashOption([]string{valAddr.String()}); err != nil {
			return nil, fmt.Errorf("module %s: %w", moduleName, err)
		}
		optsMap[moduleName] = opts
	}

	return optsMap, nil
}
