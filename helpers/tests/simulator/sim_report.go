package simulator

import (
	"fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution"

	"github.com/dfinance/dnode/cmd/config"
)

type SimReportWriter interface {
	Write(SimReportItem)
}

type SimReportItem struct {
	Index         int           // report sequential number
	BlockHeight   int64         // block height
	BlockTime     time.Time     // block time
	SimulationDur time.Duration // simulation duration
	//
	StakingBonded          sdk.Int // bonded tokens (staking pool)
	StakingNotBonded       sdk.Int // not bonded tokens (staking pool)
	RedelegationsInProcess int     // redelegations in progress
	//
	MintMinInflation     sdk.Dec // annual min inflation
	MintMaxInflation     sdk.Dec // annual max inflation
	MintAnnualProvisions sdk.Dec // annual inflation provision (not including FoundationPool)
	MintBlocksPerYear    uint64  // blocks per year estimation
	//
	DistPublicTreasuryPool     sdk.Dec // PublicTreasuryPool funds
	DistFoundationPool         sdk.Dec // FoundationPool funds
	DistLiquidityProvidersPool sdk.Dec // LiquidityProvidersPool funds
	DistHARP                   sdk.Dec // HARP funds
	//
	SupplyTotal sdk.Int // total supply
	//
	StatsBondedRatio sdk.Dec // BondedTokens / TotalSupply ratio
	//
	Counters Counter
}

// NewReportOp captures report.
func NewReportOp(period time.Duration, debug bool, writers ...SimReportWriter) *SimOperation {
	reportItemIdx := 1

	handler := func(s *Simulator) bool {
		// gather the data

		// simulation
		simBlockHeight := s.app.LastBlockHeight()
		simBlockTime := s.prevBlockTime
		_, simDur := s.SimulatedDur()
		// staking
		stakingPool := s.QueryStakePools()
		acitveRedelegations := 0
		for _, acc := range s.accounts {
			acitveRedelegations += len(s.QueryStakeRedelegations(acc.Address, sdk.ValAddress{}, sdk.ValAddress{}))
		}
		// mint
		mintParams := s.QueryMintParams()
		mintAnnualProvisions := s.QueryMintAnnualProvisions()
		mintBlocksPerYear := s.QueryMintBlocksPerYearEstimation()
		// distribution
		treasuryPool := s.QueryDistPool(distribution.PublicTreasuryPoolName)
		foundationPool := s.QueryDistPool(distribution.FoundationPoolName)
		liquidityPool := s.QueryDistPool(distribution.LiquidityProvidersPoolName)
		harpPool := s.QueryDistPool(distribution.HARPName)
		// supply
		totalSupply := s.QuerySupplyTotal()

		item := SimReportItem{
			Index:         reportItemIdx,
			BlockHeight:   simBlockHeight,
			BlockTime:     simBlockTime,
			SimulationDur: simDur,
			//
			StakingBonded:          stakingPool.BondedTokens,
			StakingNotBonded:       stakingPool.NotBondedTokens,
			RedelegationsInProcess: acitveRedelegations,
			//
			MintMinInflation:     mintParams.InflationMin,
			MintMaxInflation:     mintParams.InflationMax,
			MintAnnualProvisions: mintAnnualProvisions,
			MintBlocksPerYear:    mintBlocksPerYear,
			//
			DistPublicTreasuryPool:     treasuryPool.AmountOf(config.MainDenom),
			DistFoundationPool:         foundationPool.AmountOf(config.MainDenom),
			DistLiquidityProvidersPool: liquidityPool.AmountOf(config.MainDenom),
			DistHARP:                   harpPool.AmountOf(config.MainDenom),
			//
			SupplyTotal: totalSupply.AmountOf(config.MainDenom),
			//
			Counters: s.counter,
		}

		// calculate statistics
		item.StatsBondedRatio = sdk.NewDecFromInt(item.StakingBonded).Quo(sdk.NewDecFromInt(item.SupplyTotal))
		reportItemIdx++

		for _, writer := range writers {
			writer.Write(item)
		}

		if debug {
			debugItem := BuildDebugReportItem(s)
			fmt.Println(debugItem.String())
		}

		return true
	}

	return NewSimOperation(period, NewPeriodicNextExecFn(), handler)
}

type SimReportConsoleWriter struct {
	startedAt           time.Time
	amountDecimalsRatio sdk.Dec
}

func (w *SimReportConsoleWriter) Write(item SimReportItem) {
	reportingDur := time.Since(w.startedAt)

	str := strings.Builder{}

	str.WriteString(fmt.Sprintf("Report (%v):\n", reportingDur))
	str.WriteString(fmt.Sprintf("  BlockHeight:               %d\n", item.BlockHeight))
	str.WriteString(fmt.Sprintf("  BlockTime:                 %s\n", item.BlockTime.Format("02.01.2006T15:04:05")))
	str.WriteString(fmt.Sprintf("  SimDuration:               %v\n", FormatDuration(item.SimulationDur)))
	str.WriteString(fmt.Sprintf("   Staking: Bonded:          %s\n", w.FormatIntDecimals(item.StakingBonded)))
	str.WriteString(fmt.Sprintf("   Staking: NotBonded:       %s\n", w.FormatIntDecimals(item.StakingNotBonded)))
	str.WriteString(fmt.Sprintf("   Staking: Redelegations:   %d\n", item.RedelegationsInProcess))
	str.WriteString(fmt.Sprintf("    Mint: MinInflation:      %s\n", item.MintMinInflation))
	str.WriteString(fmt.Sprintf("    Mint: MaxInflation:      %s\n", item.MintMaxInflation))
	str.WriteString(fmt.Sprintf("    Mint: AnnualProvision:   %s\n", w.FormatDecDecimals(item.MintAnnualProvisions)))
	str.WriteString(fmt.Sprintf("    Mint: BlocksPerYear:     %d\n", item.MintBlocksPerYear))
	str.WriteString(fmt.Sprintf("     Dist: FoundationPool:   %s\n", w.FormatDecDecimals(item.DistFoundationPool)))
	str.WriteString(fmt.Sprintf("     Dist: PTreasuryPool:    %s\n", w.FormatDecDecimals(item.DistPublicTreasuryPool)))
	str.WriteString(fmt.Sprintf("     Dist: LiquidityPPool:   %s\n", w.FormatDecDecimals(item.DistLiquidityProvidersPool)))
	str.WriteString(fmt.Sprintf("     Dist: HARP:             %s\n", w.FormatDecDecimals(item.DistHARP)))
	str.WriteString(fmt.Sprintf("      Supply: Total:         %s\n", w.FormatIntDecimals(item.SupplyTotal)))
	str.WriteString(fmt.Sprintf("  Stats: Bonded/TotalSupply: %s\n", item.StatsBondedRatio))
	str.WriteString(fmt.Sprintf("  Counters:\n"))
	str.WriteString(fmt.Sprintf("    Delegations:             %d\n", item.Counters.Delegations))
	str.WriteString(fmt.Sprintf("    Redelegations:           %d\n", item.Counters.Redelegations))
	str.WriteString(fmt.Sprintf("    Undelegations:           %d\n", item.Counters.Undelegations))
	str.WriteString(fmt.Sprintf("    Rewards:                 %d\n", item.Counters.Rewards))
	str.WriteString(fmt.Sprintf("    RewardsCollected:        %s\n", w.FormatIntDecimals(item.Counters.RewardsCollected)))
	str.WriteString(fmt.Sprintf("    Commissions:             %d\n", item.Counters.Commissions))
	str.WriteString(fmt.Sprintf("    CommissionsCollected:    %s\n", w.FormatIntDecimals(item.Counters.CommissionsCollected)))

	fmt.Println(str.String())
}

func (w *SimReportConsoleWriter) FormatIntDecimals(value sdk.Int) string {
	valueDec := sdk.NewDecFromInt(value)
	fixedDec := valueDec.Mul(w.amountDecimalsRatio)

	return fixedDec.String()
}

func (w *SimReportConsoleWriter) FormatDecDecimals(value sdk.Dec) string {
	fixedDec := value.Mul(w.amountDecimalsRatio)

	return fixedDec.String()
}

func NewSimReportConsoleWriter(decimals uint8) *SimReportConsoleWriter {
	w := &SimReportConsoleWriter{
		startedAt:           time.Now(),
		amountDecimalsRatio: sdk.NewDecWithPrec(1, 0),
	}

	if decimals > 0 {
		w.amountDecimalsRatio = sdk.NewDecWithPrec(1, int64(decimals))
	}

	return w
}

// FormatDuration yet another duration formatter.
// 1.2.1 years -> 1 year, 2 months and 1 week
// 5.30 hours -> 5 hours and 30 minutes
func FormatDuration(dur time.Duration) string {
	const (
		dayDur   = 24 * time.Hour
		weekDur  = 7 * dayDur
		monthDur = 4 * weekDur
		yearDur  = 12 * monthDur
	)

	dur = dur.Round(time.Minute)

	years := dur / yearDur
	dur -= years * yearDur
	months := dur / monthDur
	dur -= months * monthDur
	weeks := dur / weekDur
	dur -= weeks * weekDur
	hours := dur / time.Hour
	dur -= hours * time.Hour
	mins := dur / time.Minute

	str := strings.Builder{}
	str.WriteString(fmt.Sprintf("%d.%d.%d years ", years, months, weeks))
	str.WriteString(fmt.Sprintf("%d.%d hours", hours, mins))

	return str.String()
}
