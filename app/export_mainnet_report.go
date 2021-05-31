package app

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/exported"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	distTypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/dfinance/dnode/helpers"
)

type (
	// MainnetBalanceReport contains initial and final (after the export) balances per accounts.
	MainnetBalanceReport map[string]*MainnetBalanceReportItem // key: account address

	MainnetBalanceReportItem struct {
		AccAddress  sdk.AccAddress    // account address
		GenCoins    sdk.Coins         // account genesis balance
		IssueCoins  sdk.Coins         // issued with Staker total coins
		AccBalance  sdk.Coins         // account current balance (after the main export routines)
		StakerEntry *StakerReportItem // Staker entry for issued coins (one per account)
		//
		IssueBondingDenom string // issued bonding coins denom
		IssueLPDenom      string // issued LP coins denom
		BondingDenom      string // final bonding coins denom
		LPDenom           string // final LP coins denom
	}

	// StakerReportItem is a parsed Staker CSV-report.
	StakerReportItem struct {
		TxHash        string         `json:"txHash"`  // Stake Tx hash
		AccAddress    sdk.AccAddress `json:"dfiAddr"` // Account address
		EthAddress    string         `json:"ethAddr"` // SrcEthereum address
		BondingAmount sdk.Int        `json:"XFI"`     // Bonding tokens stake amount
		LPAmount      sdk.Int        `json:"LPT"`     // LP tokens stake amount
		Active        bool           `json:"active"`  // true: deposited (stake is on balance) / false: withdrawn (shouldn't be on balance)
	}
)

type (
	// MainnetBalanceResults contains list of MainnetBalanceReport entries with estimated (current - initial) "loss" diffs.
	MainnetBalanceResults []MainnetBalanceResult

	MainnetBalanceResult struct {
		ReportItem  MainnetBalanceReportItem
		BondingDiff sdk.Int
		LPDiff      sdk.Int
	}
)

type (
	// MainnetBalanceStats contains MainnetBalanceResults minting stats with some extra meta.
	MainnetBalanceStats struct {
		// Total amount of bonding tokens that should be refunded [normalized decimals]
		TotalNegativeBondingDiffs sdk.Dec `json:"total_negative_bonding_diffs"`
		// Total amount of bonding tokens that all account have earned (comparing to their initial balance) [normalized decimals]
		TotalPositiveBondingDiffs sdk.Dec `json:"total_positive_bonding_diffs"`
		// Refund distribution data per account
		AccBondingMints []MainnetBalanceAccMintStats `json:"acc_bonding_mints"`
		// Sum of all current account bonding balances + total refund amount [normalized decimals]
		TotalBondingSupply sdk.Dec `json:"total_bonding_supply"`
		// Sum of all current account LP balances [normalized decimals]
		TotalLPSupply sdk.Dec `json:"total_lp_supply"`
		// Balances that are not included into current account balances (what is left after the migration) [normalized decimals]
		Leftovers MainnetBalanceLeftoversStats `json:"leftovers"`
	}

	MainnetBalanceAccMintStats struct {
		// Account address
		AccAddress sdk.AccAddress `json:"acc_address"`
		// Refund coin
		MintCoin sdk.Coin `json:"mint_coin"`
		// Refund amount [normalized decimals]
		DecAmount sdk.Dec `json:"dec_amount"`
	}

	MainnetBalanceLeftoversStats struct {
		OutstandingRewards sdk.DecCoins          `json:"outstanding_rewards,omitempty"`
		RewardPools        distTypes.RewardPools `json:"reward_pools"`
	}
)

// AppendStakerJSONReport modifies a MainnetBalanceReport with staker JSON report data.
func (r MainnetBalanceReport) AppendStakerJSONReport(filePath string) error {
	if filePath == "" {
		return nil
	}

	entriesBz, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("report open: %w", err)
	}

	entries := make([]StakerReportItem, 0)
	if err := json.Unmarshal(entriesBz, &entries); err != nil {
		return fmt.Errorf("report unmarshal: %w", err)
	}

	for i := 0; i < len(entries); i++ {
		if err := r.appendStakerReportItem(entries[i]); err != nil {
			return fmt.Errorf("entry (%d): %w", i, err)
		}
	}

	return nil
}

// AppendStakerCSVReport modifies a MainnetBalanceReport with staker CSV report data.
//
// Deprecated: Staker now exports a JSON.
func (r MainnetBalanceReport) AppendStakerCSVReport(filePath string) error {
	const (
		csvEntryColumns = 5
	)

	if filePath == "" {
		return nil
	}

	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("report open: %w", err)
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

		stakerReport := StakerReportItem{
			TxHash:        stakerTxHash,
			AccAddress:    stakerAccAddress,
			EthAddress:    stakerEthAddress,
			BondingAmount: stakerBondingAmt,
			LPAmount:      stakerLPAmt,
		}
		if err := r.appendStakerReportItem(stakerReport); err != nil {
			return fmt.Errorf("entry (%d): %w", entryIdx, err)
		}
	}

	return nil
}

// appendStakerReportItem adds Staker entry to MainnetBalanceReport.
func (r MainnetBalanceReport) appendStakerReportItem(entry StakerReportItem) error {
	reportItem, found := r[entry.AccAddress.String()]
	if !found {
		return fmt.Errorf("reportEntry for AccAddress (%s): not found", entry.AccAddress)
	}
	if reportItem.StakerEntry != nil {
		return fmt.Errorf("reportEntry for AccAddress (%s): duplicated", entry.AccAddress)
	}

	if entry.BondingAmount.IsNil() {
		entry.BondingAmount = sdk.ZeroInt()
	}
	if entry.LPAmount.IsNil() {
		entry.LPAmount = sdk.ZeroInt()
	}

	reportItem.StakerEntry = &entry

	return nil
}

// Verify compares issues data with Staker report data.
func (r MainnetBalanceReport) Verify() error {
	for accAddr, reportItem := range r {
		if reportItem.StakerEntry == nil {
			continue
		}

		issuedBondingAmt := reportItem.IssueCoins.AmountOf(reportItem.IssueBondingDenom)
		issuedLPAmt := reportItem.IssueCoins.AmountOf(reportItem.IssueLPDenom)
		stakerBondingAmt := reportItem.StakerEntry.BondingAmount
		stakerLPAmt := reportItem.StakerEntry.LPAmount

		if !issuedBondingAmt.Equal(stakerBondingAmt) {
			return fmt.Errorf("account (%s): issued / staker Bonding amount mismatch: %s / %s", accAddr, issuedBondingAmt, stakerBondingAmt)
		}
		if !issuedLPAmt.Equal(stakerLPAmt) {
			return fmt.Errorf("account (%s): issued / staker LP amount mismatch: %s / %s", accAddr, issuedLPAmt, stakerLPAmt)
		}
	}

	return nil
}

// GetResults builds MainnetBalanceResults report.
func (r MainnetBalanceReport) GetResults() MainnetBalanceResults {
	results := make(MainnetBalanceResults, 0, len(r))
	for _, reportItem := range r {
		diffBonding := reportItem.GetCurrentBondingBalance().Sub(reportItem.GetInitialBondingBalance())
		diffLP := reportItem.GetCurrentLPBalance().Sub(reportItem.GetInitialLPBalance())

		results = append(results, MainnetBalanceResult{
			ReportItem:  *reportItem,
			BondingDiff: diffBonding,
			LPDiff:      diffLP,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].ReportItem.GetCurrentBondingBalance().GTE(results[j].ReportItem.GetCurrentBondingBalance())
	})

	return results
}

// GetInitialBondingBalance returns initial amount for BondingDenom (gen balance + issues).
func (i MainnetBalanceReportItem) GetInitialBondingBalance() sdk.Int {
	genAmt := i.GenCoins.AmountOf(i.BondingDenom)
	issuedAmt := i.IssueCoins.AmountOf(i.IssueBondingDenom)
	if i.StakerEntry != nil && !i.StakerEntry.Active {
		issuedAmt = issuedAmt.Sub(i.StakerEntry.BondingAmount)
	}

	return genAmt.Add(issuedAmt)
}

// GetInitialLPBalance returns initial amount for LPDenom (gen balance + issues).
func (i MainnetBalanceReportItem) GetInitialLPBalance() sdk.Int {
	genAmt := i.GenCoins.AmountOf(i.LPDenom)
	issuedAmt := i.IssueCoins.AmountOf(i.IssueLPDenom)
	if i.StakerEntry != nil && !i.StakerEntry.Active {
		issuedAmt = issuedAmt.Sub(i.StakerEntry.LPAmount)
	}

	return genAmt.Add(issuedAmt)
}

// GetCurrentBondingBalance returns final amount for BondingDenom.
func (i MainnetBalanceReportItem) GetCurrentBondingBalance() sdk.Int {
	return i.AccBalance.AmountOf(i.BondingDenom)
}

// GetCurrentLPBalance returns final amount for LPDenom.
func (i MainnetBalanceReportItem) GetCurrentLPBalance() sdk.Int {
	return i.AccBalance.AmountOf(i.LPDenom)
}

// SaveToCSV saves MainnetBalanceResults to a CSV file.
func (r MainnetBalanceResults) SaveToCSV(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}
	defer f.Close()
	csvWriter := csv.NewWriter(f)

	// Header
	err = csvWriter.Write([]string{
		"AccAddress",
		"Genesis balance [xfi]",
		"Genesis balance [lp]",
		"Issued balance [xfi]",
		"Issued balance [lp]",
		"Current balance [xfi]",
		"Current balance [lp]",
		"Balance diff [xfi]",
		"Balance diff [lp]",
	})
	if err != nil {
		return fmt.Errorf("header write: %w", err)
	}

	// Entries
	for i, result := range r {
		issueBondDenom, issueLPDenom := result.ReportItem.IssueBondingDenom, result.ReportItem.IssueLPDenom
		bondDenom, lpDenom := result.ReportItem.BondingDenom, result.ReportItem.LPDenom

		err := csvWriter.Write([]string{
			result.ReportItem.AccAddress.String(),
			result.ReportItem.GenCoins.AmountOf(bondDenom).String(),
			result.ReportItem.GenCoins.AmountOf(lpDenom).String(),
			result.ReportItem.IssueCoins.AmountOf(issueBondDenom).String(),
			result.ReportItem.IssueCoins.AmountOf(issueLPDenom).String(),
			result.ReportItem.AccBalance.AmountOf(bondDenom).String(),
			result.ReportItem.AccBalance.AmountOf(lpDenom).String(),
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

// SaveToJSON saves MainnetBalanceStats to a JSON file.
func (s MainnetBalanceStats) SaveToJSON(path string) error {
	statsBz, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}
	defer f.Close()

	if _, err := f.Write(statsBz); err != nil {
		return fmt.Errorf("write to file: %w", err)
	}

	return nil
}

// GetTotalMintCoins returns sdk.Coins that need to be minted.
func (s MainnetBalanceStats) GetTotalMintCoins() sdk.Coins {
	coins := sdk.NewCoins()
	for _, acc := range s.AccBondingMints {
		coins = coins.Add(acc.MintCoin)
	}

	return coins
}

func NewMainnetBalanceReport(ctx sdk.Context, app *DnServiceApp,
	issueBondingDenom, issueLPDenom, bondingDenom, lpDenom string,
	stakerCSVReportPath string,
) (MainnetBalanceReport, error) {

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
			AccAddress:     "wallet19xshddf5ww7fhd53fumly2r7lqsszz63fxca9x",
			BondingBalance: "2500000000000000000000",
			LPBalance:      "0",
		},
		{
			AccAddress:     "wallet1l9mukqvh0etam66dvgw99w9awv3jjv6tyh2hpc",
			BondingBalance: "50000000000000000000000",
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
	}

	report := make(MainnetBalanceReport)

	// iterate over all accounts and fill the report items
	{
		app.accountKeeper.IterateAccounts(ctx, func(acc exported.Account) (stop bool) {
			report[acc.GetAddress().String()] = &MainnetBalanceReportItem{
				AccAddress:        acc.GetAddress(),
				GenCoins:          sdk.NewCoins(),
				IssueCoins:        sdk.NewCoins(),
				AccBalance:        acc.GetCoins(),
				StakerEntry:       nil,
				IssueBondingDenom: issueBondingDenom,
				IssueLPDenom:      issueLPDenom,
				BondingDenom:      bondingDenom,
				LPDenom:           lpDenom,
			}
			return false
		})
	}

	// set genesis balances
	{
		for i, genBalance := range genBalances {
			accAddress, err := sdk.AccAddressFromBech32(genBalance.AccAddress)
			if err != nil {
				return nil, fmt.Errorf("genBalance (%d): AccAddress (%s): invalid: %w", i, genBalance.AccAddress, err)
			}
			bondingAmt, ok := sdk.NewIntFromString(genBalance.BondingBalance)
			if !ok {
				return nil, fmt.Errorf("genBalance (%d): BondingBalance (%s): invalid sdk.Int", i, genBalance.BondingBalance)
			}
			lpAmt, ok := sdk.NewIntFromString(genBalance.LPBalance)
			if !ok {
				return nil, fmt.Errorf("genBalance (%d): LPBalance (%s): invalid sdk.Int", i, genBalance.LPBalance)
			}

			reportItem, found := report[accAddress.String()]
			if !found {
				return nil, fmt.Errorf("genBalance (%d): account (%s): not found", i, accAddress)
			}
			reportItem.GenCoins = sdk.NewCoins(
				sdk.NewCoin(bondingDenom, bondingAmt),
				sdk.NewCoin(lpDenom, lpAmt),
			)
		}
	}

	// iterate over all issues and aggregate duplicate issues
	for _, issue := range app.ccKeeper.GetGenesisIssues(ctx) {
		accAddr := issue.Payee

		reportItem, found := report[accAddr.String()]
		if !found {
			return nil, fmt.Errorf("issue (%s): account (%s): not found", issue.ID, accAddr)
		}
		reportItem.IssueCoins = reportItem.IssueCoins.Add(issue.Coin)
	}

	// update report with Staker CSV-report
	if err := report.AppendStakerJSONReport(stakerCSVReportPath); err != nil {
		return nil, fmt.Errorf("AppendStakerJSONReport: %w", err)
	}

	return report, nil
}

func NewMainnetBalanceStats(ctx sdk.Context, app *DnServiceApp, results MainnetBalanceResults) MainnetBalanceStats {
	decimalDec := sdk.NewDecWithPrec(1, 18)
	normalizeInt := func(v sdk.Int) sdk.Dec {
		return v.ToDec().Mul(decimalDec)
	}
	normalizeDec := func(v sdk.Dec) sdk.Dec {
		return v.Mul(decimalDec)
	}
	normalizeDecCoins := func(v sdk.DecCoins) sdk.DecCoins {
		ret := make(sdk.DecCoins, 0, len(v))
		for _, coin := range v {
			coin.Amount = normalizeDec(coin.Amount)
			ret = append(ret, coin)
		}
		return ret
	}

	// calculate total supply, total diffs and distribution amount per account
	totalBonding, totalLP := sdk.ZeroInt(), sdk.ZeroInt()
	totalPositiveDiffs, totalNegativeDiffs := sdk.ZeroInt(), sdk.ZeroInt()
	accMints := make([]MainnetBalanceAccMintStats, 0, len(results))

	for _, result := range results {
		totalBonding = totalBonding.Add(result.ReportItem.GetCurrentBondingBalance())
		totalLP = totalLP.Add(result.ReportItem.GetCurrentLPBalance())

		if !result.BondingDiff.IsNegative() {
			totalPositiveDiffs = totalPositiveDiffs.Add(result.BondingDiff)
		} else {
			totalNegativeDiffs = totalNegativeDiffs.Add(result.BondingDiff)

			mintCoin := sdk.NewCoin(result.ReportItem.BondingDenom, result.BondingDiff.MulRaw(-1))
			accMints = append(accMints, MainnetBalanceAccMintStats{
				AccAddress: result.ReportItem.AccAddress,
				MintCoin:   mintCoin,
				DecAmount:  normalizeInt(mintCoin.Amount),
			})
		}
	}
	totalBonding = totalBonding.Add(totalNegativeDiffs.MulRaw(-1))

	sort.Slice(accMints, func(i, j int) bool {
		return accMints[i].MintCoin.Amount.GTE(accMints[j].MintCoin.Amount)
	})

	// leftovers
	leftoverOutstanding := sdk.NewDecCoins()
	app.distrKeeper.IterateValidatorOutstandingRewards(ctx, func(_ sdk.ValAddress, rewards types.ValidatorOutstandingRewards) (stop bool) {
		leftoverOutstanding = leftoverOutstanding.Add(rewards...)
		return false
	})

	rewardPools := app.distrKeeper.GetRewardPools(ctx)

	return MainnetBalanceStats{
		TotalNegativeBondingDiffs: normalizeInt(totalNegativeDiffs),
		TotalPositiveBondingDiffs: normalizeInt(totalPositiveDiffs),
		AccBondingMints:           accMints,
		TotalBondingSupply:        normalizeInt(totalBonding),
		TotalLPSupply:             normalizeInt(totalLP),
		Leftovers: MainnetBalanceLeftoversStats{
			OutstandingRewards: normalizeDecCoins(leftoverOutstanding),
			RewardPools: distTypes.RewardPools{
				LiquidityProvidersPool: normalizeDecCoins(rewardPools.LiquidityProvidersPool),
				FoundationPool:         normalizeDecCoins(rewardPools.FoundationPool),
				PublicTreasuryPool:     normalizeDecCoins(rewardPools.PublicTreasuryPool),
				HARP:                   normalizeDecCoins(rewardPools.HARP),
			},
		},
	}
}
