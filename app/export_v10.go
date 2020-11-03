package app

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/staking/exported"
	"github.com/cosmos/cosmos-sdk/x/supply"

	"github.com/dfinance/dnode/cmd/config/genesis/defaults"
	"github.com/dfinance/dnode/helpers"
	"github.com/dfinance/dnode/x/ccstorage"
	"github.com/dfinance/dnode/x/vmauth"
)

// setMainnetZeroHeightOptionsV10 updates options map per module for Testnet v0.7 -> Mainnet v1.0 migration.
// Options removes all XFI tokens and renames SXFI -> XFI.
func setMainnetZeroHeightOptionsV10(optsMap map[string]interface{}) (map[string]interface{}, error) {
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

// SXFIBalanceReportItem keeps initial, staked and reward balances per account.
type SXFIBalanceReportItem struct {
	AccAddress        sdk.AccAddress
	AccBalance        sdk.Coins
	IssueCoins        sdk.Coins
	RewardCoins       sdk.Coins
	DelBondingShares  sdk.Dec
	DelLPShares       sdk.Dec
	DelBondingTokens  sdk.Dec
	DelLPTokens       sdk.Dec
	GenCoins          sdk.Coins
	StakerReport      *SXFIStakerReportItem
	IssueBondingDenom string
	IssueLPDenom      string
	BondingDenom      string
	LPDenom           string
}

// SXFIStakerReportItem is a parsed Staker CSV-report.
type SXFIStakerReportItem struct {
	TxHash        string
	AccAddress    sdk.AccAddress
	EthAddress    string
	BondingAmount sdk.Int
	LPAmount      sdk.Int
}

// GetInitialBondingBalance returns initial amount for BondingDenom (gen balance + issues).
func (i SXFIBalanceReportItem) GetInitialBondingBalance() sdk.Int {
	genAmt := i.GenCoins.AmountOf(i.BondingDenom)
	issuedAmt := i.IssueCoins.AmountOf(i.IssueBondingDenom)
	if i.StakerReport == nil {
		issuedAmt = sdk.ZeroInt()
	}

	return genAmt.Add(issuedAmt)
}

// GetIssueBondingBalance returns initial amount for LPDenom (gen balance + issues).
func (i SXFIBalanceReportItem) GetInitialLPBalance() sdk.Int {
	genAmt := i.GenCoins.AmountOf(i.LPDenom)
	issuedAmt := i.IssueCoins.AmountOf(i.IssueLPDenom)
	if i.StakerReport == nil {
		issuedAmt = sdk.ZeroInt()
	}

	return genAmt.Add(issuedAmt)
}

// GetCurrentBondingBalance returns final amount for BondingDenom (current balance + rewards + delegations).
func (i SXFIBalanceReportItem) GetCurrentBondingBalance() sdk.Int {
	accBalanceAmt := i.AccBalance.AmountOf(i.BondingDenom)
	rewardAmt := i.RewardCoins.AmountOf(i.BondingDenom)
	delAmt := i.DelBondingTokens.TruncateInt()

	return accBalanceAmt.Add(rewardAmt).Add(delAmt)
}

// GetCurrentLPBalance returns final amount for LPDenom (current balance + rewards + delegations).
func (i SXFIBalanceReportItem) GetCurrentLPBalance() sdk.Int {
	accBalanceAmt := i.AccBalance.AmountOf(i.LPDenom)
	rewardAmt := i.RewardCoins.AmountOf(i.LPDenom)
	delAmt := i.DelLPTokens.TruncateInt()

	return accBalanceAmt.Add(rewardAmt).Add(delAmt)
}

func NewSXFIBalanceReportItem(accAddr sdk.AccAddress, accCoins sdk.Coins, issueBondingDenom, issueLPDenom, bondingDenom, lpDenom string) *SXFIBalanceReportItem {
	return &SXFIBalanceReportItem{
		AccAddress:        accAddr,
		AccBalance:        accCoins,
		IssueCoins:        sdk.NewCoins(),
		RewardCoins:       sdk.NewCoins(),
		DelBondingShares:  sdk.ZeroDec(),
		DelLPShares:       sdk.ZeroDec(),
		DelBondingTokens:  sdk.ZeroDec(),
		DelLPTokens:       sdk.ZeroDec(),
		GenCoins:          sdk.NewCoins(),
		StakerReport:      nil,
		IssueBondingDenom: issueBondingDenom,
		IssueLPDenom:      issueLPDenom,
		BondingDenom:      bondingDenom,
		LPDenom:           lpDenom,
	}
}

type SXFIBalanceReportResult struct {
	ReportItem  SXFIBalanceReportItem
	BondingDiff sdk.Int
	LPDiff      sdk.Int
}

type SXFIBalanceReportResults []SXFIBalanceReportResult

func (results SXFIBalanceReportResults) SaveToCSV(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}
	defer f.Close()
	csvWriter := csv.NewWriter(f)

	// Header
	err = csvWriter.Write([]string{
		"AccAddress",
		"GenCoins",
		"IssueCoins",
		"WalletCoins",
		"RewardCoins",
		"DelBondingTokens",
		"DelLPTokens",
		"BondingDiff",
		"LPDiff",
	})
	if err != nil {
		return fmt.Errorf("header write: %w", err)
	}

	// Entries
	for i, result := range results {
		err := csvWriter.Write([]string{
			result.ReportItem.AccAddress.String(),
			result.ReportItem.GenCoins.String(),
			result.ReportItem.IssueCoins.String(),
			result.ReportItem.AccBalance.String(),
			result.ReportItem.RewardCoins.String(),
			result.ReportItem.DelBondingTokens.String(),
			result.ReportItem.DelLPTokens.String(),
			result.BondingDiff.String(),
			result.LPDiff.String(),
		})
		if err != nil {
			return fmt.Errorf("entry %d: write: %w", i+1, err)
		}
	}

	csvWriter.Flush()

	return nil
}

func (results SXFIBalanceReportResults) String() string {
	decimalDec := sdk.NewDecWithPrec(1, 18)

	str := strings.Builder{}
	str.WriteString("Mainnet SXFI-XFI relation report:\n")
	for _, result := range results {
		diffBondingDec, diffLPDec := result.BondingDiff.ToDec().Mul(decimalDec), result.LPDiff.ToDec().Mul(decimalDec)
		str.WriteString(fmt.Sprintf("  - %s\n", result.ReportItem.AccAddress))
		str.WriteString(fmt.Sprintf("    BondingDiff: %s (%s)\n", diffBondingDec, result.BondingDiff))
		str.WriteString(fmt.Sprintf("    LPDiff:      %s (%s)\n", diffLPDec, result.LPDiff))
		str.WriteString(fmt.Sprintf("    GenBalance:  %s\n", result.ReportItem.GenCoins))
		str.WriteString(fmt.Sprintf("    AccBalance:  %s\n", result.ReportItem.AccBalance))
		str.WriteString(fmt.Sprintf("    Issues:      %s\n", result.ReportItem.IssueCoins))
		str.WriteString(fmt.Sprintf("    Rewards:     %s\n", result.ReportItem.RewardCoins))
		str.WriteString(fmt.Sprintf("    BDel:        %s (%s)\n", result.ReportItem.DelBondingTokens, result.ReportItem.DelBondingShares))
		str.WriteString(fmt.Sprintf("    LPDel:       %s (%s)\n", result.ReportItem.DelLPTokens, result.ReportItem.DelLPShares))
	}

	return str.String()
}

// SXFIBalanceReport contains initial and final Testnet (v0.7) sxfi balance for accounts.
// Key - account address.
type SXFIBalanceReport map[string]*SXFIBalanceReportItem

// AppendGenesisBalances modifies a SXFIBalanceReport with genesis account balances.
func (r SXFIBalanceReport) AppendGenesisBalances(
	ctx sdk.Context, app *DnServiceApp,
	issueBondingDenom, issueLPDenom, bondingDenom, lpDenom string,
) error {

	genBalances := []struct {
		AccAddress     string
		BondingBalance string
		LPBalance      string
	}{
		{
			AccAddress:     "wallet1wwmenr38hhrem2v3ue3gwdhj03ynzcvlxgc92u",
			BondingBalance: "3400000000000000000000000",
			LPBalance:      "0",
		},
		{
			AccAddress:     "wallet1a6sd0y8l0ma0gnytacrnwlmnupm7ftnwxngalr",
			BondingBalance: "2500000000000000000000",
			LPBalance:      "0",
		},
		{
			AccAddress:     "wallet1whpkntyj549f7euftgpng24k2we8legght4rzg",
			BondingBalance: "2500000000000000000000",
			LPBalance:      "0",
		},
		{
			AccAddress:     "wallet1zwkqfm2sdgyx0g6h2dj9em4z4kjgy5lmtnmgjd",
			BondingBalance: "2500000000000000000000",
			LPBalance:      "0",
		},
		{
			AccAddress:     "wallet10a24shxzjtutj637rr8shwkwaxx8paplu4vc6f",
			BondingBalance: "2500000000000000000000",
			LPBalance:      "0",
		},
		{
			AccAddress:     "wallet19xshddf5ww7fhd53fumly2r7lqsszz63fxca9x",
			BondingBalance: "2500000000000000000000",
			LPBalance:      "0",
		},
		{
			AccAddress:     "wallet1l9mukqvh0etam66dvgw99w9awv3jjv6tyh2hpc",
			BondingBalance: "50000000000000000000000",
			LPBalance:      "0",
		},
	}

	for i, genBalance := range genBalances {
		accAddress, err := sdk.AccAddressFromBech32(genBalance.AccAddress)
		if err != nil {
			return fmt.Errorf("genBalance (%d): AccAddress (%s): invalid: %w", i, genBalance.AccAddress, err)
		}
		bondingAmt, ok := sdk.NewIntFromString(genBalance.BondingBalance)
		if !ok {
			return fmt.Errorf("genBalance (%d): BondingBalance (%s): invalid sdk.Int", i, genBalance.BondingBalance)
		}
		lpAmt, ok := sdk.NewIntFromString(genBalance.LPBalance)
		if !ok {
			return fmt.Errorf("genBalance (%d): LPBalance (%s): invalid sdk.Int", i, genBalance.LPBalance)
		}
		acc := app.accountKeeper.GetAccount(ctx, accAddress)
		if acc == nil {
			return fmt.Errorf("genBalance (%d): account (%s): not found", i, accAddress)
		}

		reportItem := NewSXFIBalanceReportItem(accAddress, acc.GetCoins(), issueBondingDenom, issueLPDenom, bondingDenom, lpDenom)
		reportItem.GenCoins = sdk.NewCoins(
			sdk.NewCoin(bondingDenom, bondingAmt),
			sdk.NewCoin(lpDenom, lpAmt),
		)
		r[accAddress.String()] = reportItem
	}

	return nil
}

// AppendStakerCSVReport modifies a SXFIBalanceReport with staker CSV report data.
func (r SXFIBalanceReport) AppendStakerCSVReport(filePath string) error {
	const (
		csvEntryColumns = 5
	)

	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("CSV staker report open: %w", err)
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	entryIdx := 0
	for {
		entryIdx++
		csvEntry, err := csvReader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("entry (%d): read failed: %w", entryIdx, err)
		}
		if entryIdx == 1 {
			// skip the header
			continue
		}

		// parse
		if len(csvEntry) != csvEntryColumns {
			return fmt.Errorf("entry (%d): invalid number of columns: %d / %d", entryIdx, len(csvEntry), csvEntryColumns)
		}
		stakerTxHash := csvEntry[0]
		if stakerTxHash == "" {
			return fmt.Errorf("entry (%d): TxHash: emtpy", entryIdx)
		}
		stakerBondingAmt := sdk.ZeroInt()
		if amtRaw := csvEntry[1]; amtRaw != "" {
			amt, ok := sdk.NewIntFromString(amtRaw)
			if !ok {
				return fmt.Errorf("entry (%d): BondingAmount (%s): invalid sdk.Int", entryIdx, amtRaw)
			}
			stakerBondingAmt = amt
		}
		stakerAccAddress, err := sdk.AccAddressFromBech32(csvEntry[2])
		if err != nil {
			return fmt.Errorf("entry (%d): AccAddress (%s): invalid sdk.AccAddress: %w", entryIdx, csvEntry[2], err)
		}
		stakerEthAddress := csvEntry[3]
		if !helpers.IsEthereumAddress(stakerEthAddress) {
			return fmt.Errorf("entry (%d): EthAddress (%s): invalid", entryIdx, stakerEthAddress)
		}
		stakerLPAmt := sdk.ZeroInt()
		if amtRaw := csvEntry[4]; amtRaw != "" {
			amt, ok := sdk.NewIntFromString(amtRaw)
			if !ok {
				return fmt.Errorf("entry (%d): LPAmount (%s): invalid sdk.Int", entryIdx, amtRaw)
			}
			stakerLPAmt = amt
		}

		stakerReport := &SXFIStakerReportItem{
			TxHash:        stakerTxHash,
			AccAddress:    stakerAccAddress,
			EthAddress:    stakerEthAddress,
			BondingAmount: stakerBondingAmt,
			LPAmount:      stakerLPAmt,
		}
		reportItem, found := r[stakerReport.AccAddress.String()]
		if !found {
			return fmt.Errorf("entry (%d): reportEntry for AccAddress %s: not found", entryIdx, stakerReport.AccAddress)
		}
		if reportItem.StakerReport != nil {
			return fmt.Errorf("entry (%d): reportEntry for AccAddress %s: StakerReport already exists", entryIdx, stakerReport.AccAddress)
		}

		reportItem.StakerReport = stakerReport
	}

	return nil
}

// Verify compares issues data with Staker report data.
func (r SXFIBalanceReport) Verify() error {
	for accAddr, reportItem := range r {
		if reportItem.StakerReport == nil {
			continue
		}

		issuedBondingAmt := reportItem.IssueCoins.AmountOf(reportItem.IssueBondingDenom)
		stakerBondingAmt := reportItem.StakerReport.BondingAmount
		if !issuedBondingAmt.Equal(stakerBondingAmt) {
			return fmt.Errorf("account %s: issued / staker Bonding amount mismatch: %s / %s", accAddr, issuedBondingAmt, stakerBondingAmt)
		}
	}

	return nil
}

func (r SXFIBalanceReport) GetResults() SXFIBalanceReportResults {
	results := make(SXFIBalanceReportResults, 0, len(r))
	for _, reportItem := range r {
		diffBonding := reportItem.GetCurrentBondingBalance().Sub(reportItem.GetInitialBondingBalance())
		diffLP := reportItem.GetCurrentLPBalance().Sub(reportItem.GetInitialLPBalance())

		results = append(results, SXFIBalanceReportResult{
			ReportItem:  *reportItem,
			BondingDiff: diffBonding,
			LPDiff:      diffLP,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].BondingDiff.LT(results[j].BondingDiff)
	})

	return results
}

// getMainnetSXFIBalanceReport returns a SXFIBalanceReport report.
func (app *DnServiceApp) getMainnetSXFIBalanceReport(ctx sdk.Context,
	issueBondingDenom, issueLPDenom, bondingDenom, lpDenom string,
	stakerCSVReportPath string,
) (SXFIBalanceReport, error) {

	cacheCtx, _ := ctx.CacheContext()

	// initialize report with genesis data
	report := make(SXFIBalanceReport)
	if err := report.AppendGenesisBalances(ctx, app, issueBondingDenom, issueLPDenom, bondingDenom, lpDenom); err != nil {
		return nil, fmt.Errorf("append genesis balances: %w", err)
	}

	// iterate all issues and combine duplicate payees
	for _, issue := range app.ccKeeper.GetGenesisIssues(cacheCtx) {
		accAddr := issue.Payee

		reportItem, found := report[accAddr.String()]
		if !found {
			acc := app.accountKeeper.GetAccount(cacheCtx, accAddr)
			if acc == nil {
				return nil, fmt.Errorf("issue %s: getAccount for %s: not found", issue.ID, accAddr)
			}

			reportItem = NewSXFIBalanceReportItem(accAddr, acc.GetCoins(), issueBondingDenom, issueLPDenom, bondingDenom, lpDenom)
		}

		reportItem.IssueCoins = reportItem.IssueCoins.Add(issue.Coin)
		report[accAddr.String()] = reportItem
	}

	// withdraw all rewards
	// as all rewards were transferred to rewards bank before, we only query the bank coins for each validator
	for _, reportItem := range report {
		accAddr := reportItem.AccAddress
		app.distrKeeper.IterateDelegatorRewardsBankCoins(ctx, accAddr, func(_ sdk.ValAddress, coins sdk.Coins) (stop bool) {
			reportItem.RewardCoins = reportItem.RewardCoins.Add(coins...)
			return false
		})
	}

	// unbond all delegations
	// no actual undelegation is done, we just calculate delegator tokens based on shares and validator tokens
	{
		for _, reportItem := range report {
			accAddr := reportItem.AccAddress
			var iterationErr error
			app.stakingKeeper.IterateDelegations(
				cacheCtx, accAddr,
				func(_ int64, del exported.DelegationI) (stop bool) {
					val, found := app.stakingKeeper.GetValidator(cacheCtx, del.GetValidatorAddr())
					if !found {
						iterationErr = fmt.Errorf("account %s: get delegation validator %s: not found", accAddr, del.GetValidatorAddr())
						return true
					}

					reportItem.DelBondingShares = reportItem.DelBondingShares.Add(del.GetBondingShares())
					if !del.GetBondingShares().IsZero() {
						reportItem.DelBondingTokens = reportItem.DelBondingTokens.Add(val.BondingTokensFromSharesTruncated(del.GetBondingShares()))
					}
					reportItem.DelLPShares = reportItem.DelLPShares.Add(del.GetLPShares())
					if !del.GetLPShares().IsZero() {
						reportItem.DelLPTokens = reportItem.DelLPTokens.Add(val.LPTokensFromSharesTruncated(del.GetLPShares()))
					}

					return false
				},
			)
			if iterationErr != nil {
				return nil, iterationErr
			}
		}
	}

	// update report with Staker CSV-report
	if stakerCSVReportPath != "" {
		if err := report.AppendStakerCSVReport(stakerCSVReportPath); err != nil {
			return nil, fmt.Errorf("append append StakerCSVReport: %w", err)
		}
	}

	return report, nil
}

type SXFIBalanceReportStats struct {
	TotalNegativeBondingDiffs sdk.Dec
	TotalPositiveBondingDiffs sdk.Dec
	AccMints                  map[string]sdk.DecCoin
}

// processMainnetSXFIBalance builds getMainnetSXFIBalanceReport and mints and transfers negative diffs.
func (app *DnServiceApp) processMainnetSXFIBalance(ctx sdk.Context) error {
	const (
		issueDenom   = "sxfi"
		bondingDenom = "xfi"
		lpDenom      = "lpt"
	)
	decimalDec := sdk.NewDecWithPrec(1, 18)

	stakerReportPath := os.Getenv("DN_ZHP_STAKERREPORT_PATH")
	reportOutputPrefix := os.Getenv("DN_ZHP_REPORTOUTPUT_PREFIX")
	if stakerReportPath == "" {
		return fmt.Errorf("envVar %q: not set", "DN_ZHP_STAKERREPORT_PATH")
	}
	if reportOutputPrefix == "" {
		return fmt.Errorf("envVar %q: not set", "DN_ZHP_REPORTOUTPUT_PREFIX")
	}

	// build report
	report, err := app.getMainnetSXFIBalanceReport(
		ctx,
		issueDenom, lpDenom, bondingDenom, lpDenom,
		stakerReportPath,
	)
	if err != nil {
		return fmt.Errorf("getMainnetSXFIBalanceReport: %w", err)
	}
	if err := report.Verify(); err != nil {
		return fmt.Errorf("report verification: %w", err)
	}

	// save results
	results := report.GetResults()
	if err := results.SaveToCSV(reportOutputPrefix + "data.csv"); err != nil {
		return fmt.Errorf("saving report results to CSV: %w", err)
	}

	// calculate the mint amount
	positiveDiffs, negativeDiffs := sdk.ZeroInt(), sdk.ZeroInt()
	stats := SXFIBalanceReportStats{
		TotalNegativeBondingDiffs: sdk.ZeroDec(),
		TotalPositiveBondingDiffs: sdk.ZeroDec(),
		AccMints:                  make(map[string]sdk.DecCoin, len(report)),
	}
	for _, result := range results {
		if !result.BondingDiff.IsNegative() {
			positiveDiffs = positiveDiffs.Add(result.BondingDiff)
			continue
		}
		negativeDiffs = negativeDiffs.Add(result.BondingDiff)
	}
	negativeDiffs = negativeDiffs.MulRaw(-1)
	bondingMintCoin := sdk.NewCoin(bondingDenom, negativeDiffs)
	//
	stats.TotalPositiveBondingDiffs = positiveDiffs.ToDec().Mul(decimalDec)
	stats.TotalNegativeBondingDiffs = negativeDiffs.ToDec().Mul(decimalDec)

	// mint
	if err := app.mintKeeper.MintCoins(ctx, sdk.NewCoins(bondingMintCoin)); err != nil {
		return fmt.Errorf("minting bonding coins: %w", err)
	}
	if err := app.ccsKeeper.IncreaseCurrencySupply(ctx, bondingMintCoin); err != nil {
		return fmt.Errorf("increasing ccStorage supply: %w", err)
	}

	// distribute minted coins
	for _, result := range results {
		diff := result.BondingDiff
		if !diff.IsNegative() {
			continue
		}

		coin := sdk.NewCoin(bondingDenom, diff.MulRaw(-1))
		if err := app.supplyKeeper.SendCoinsFromModuleToAccount(ctx, mint.ModuleName, result.ReportItem.AccAddress, sdk.NewCoins(coin)); err != nil {
			return fmt.Errorf("sending minted coins to %s: %w", result.ReportItem.AccAddress, err)
		}
		//
		stats.AccMints[result.ReportItem.AccAddress.String()] = sdk.NewDecCoinFromDec(
			coin.Denom,
			coin.Amount.ToDec().Mul(decimalDec),
		)
	}

	// save stats
	statsBz, err := json.Marshal(stats)
	if err != nil {
		return fmt.Errorf("stats: JSON marshal: %w", err)
	}
	f, err := os.Create(reportOutputPrefix + "stats.json")
	if err != nil {
		return fmt.Errorf("stats: creating file: %w", err)
	}
	defer f.Close()
	if _, err := f.Write(statsBz); err != nil {
		return fmt.Errorf("stats: write to file: %w", err)
	}

	// check the invariants
	if err := app.checkInvariants(ctx); err != nil {
		return fmt.Errorf("post invariants check: %w", err)
	}

	return nil
}
