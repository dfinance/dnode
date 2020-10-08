package simulator

import (
	"fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution"
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
	ValidatorsBonded    int // number of bonded validators
	ValidatorsUnbonding int // number of unbonding validators
	ValidatorsUnbonded  int // number of unbonded validators
	//
	StakingBonded          sdk.Int // bonded tokens (staking pool)
	StakingNotBonded       sdk.Int // not bonded tokens (staking pool)
	StakingLPs             sdk.Int // bonded LP tokens (LPs pool)
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
	DistModuleBalanceMain      sdk.Int // Distribution module balance [main]
	DistModuleBalanceStaking   sdk.Int // Distribution module balance [staking]
	DistBankBalanceMain        sdk.Int // Distribution bank balance [main]
	DistBankBalanceStaking     sdk.Int // Distribution bank balance [staking]
	DistLockedRatio            sdk.Dec // Current locked ratio
	//
	SupplyTotalMain    sdk.Int // total supply [main denom]
	SupplyTotalStaking sdk.Int // total supply [staking denom]
	SupplyTotalLP      sdk.Int // total supply [LP denom]
	//
	StatsBondedRatio sdk.Dec // BondedTokens / TotalSupply ratio [staking denom]
	StatsLPRatio     sdk.Dec // BondedTokens / TotalSupply ratio [LP denom]
	//
	AccsBalanceMain    sdk.Int
	AccsBalanceStaking sdk.Int
	//
	Counters Counter
	//
	formatIntDecimals   func(value sdk.Int) string
	formatDecDecimals   func(value sdk.Dec) string
	formatCoinsDecimals func(coins sdk.Coins) string
}

// NewReportOp captures report.
func NewReportOp(period time.Duration, debug bool, writers ...SimReportWriter) *SimOperation {
	id := "ReportOp"
	reportItemIdx := 1

	handler := func(s *Simulator) (bool, string) {
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
		// validators
		bondedCnt, unbondingCnt, unbondedCnt := 0, 0, 0
		for _, val := range s.QueryReadAllValidators() {
			if val.IsBonded() {
				bondedCnt++
				continue
			}
			if val.IsUnbonding() {
				unbondingCnt++
				continue
			}
			unbondedCnt++
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
		distrMAccBalance := s.QuerySupplyModuleBalance(distribution.ModuleName)
		distrBankBalance := s.QuerySupplyModuleBalance(distribution.RewardsBankPoolName)

		// supply
		totalSupply := s.QuerySupplyTotal()

		// accounts
		accsTotalMain, accsTotalStaking := sdk.ZeroInt(), sdk.ZeroInt()
		for _, acc := range s.GetAllAccounts() {
			accsTotalMain = accsTotalMain.Add(acc.Coins.AmountOf(s.mainDenom))
			accsTotalStaking = accsTotalStaking.Add(acc.Coins.AmountOf(s.stakingDenom))
		}

		item := SimReportItem{
			Index:         reportItemIdx,
			BlockHeight:   simBlockHeight,
			BlockTime:     simBlockTime,
			SimulationDur: simDur,
			//
			ValidatorsBonded:    bondedCnt,
			ValidatorsUnbonding: unbondingCnt,
			ValidatorsUnbonded:  unbondedCnt,
			//
			StakingBonded:          stakingPool.BondedTokens,
			StakingNotBonded:       stakingPool.NotBondedTokens,
			StakingLPs:             stakingPool.LiquidityTokens,
			RedelegationsInProcess: acitveRedelegations,
			//
			MintMinInflation:     mintParams.InflationMin,
			MintMaxInflation:     mintParams.InflationMax,
			MintAnnualProvisions: mintAnnualProvisions,
			MintBlocksPerYear:    mintBlocksPerYear,
			//
			DistPublicTreasuryPool:     treasuryPool.AmountOf(s.stakingDenom),
			DistFoundationPool:         foundationPool.AmountOf(s.stakingDenom),
			DistLiquidityProvidersPool: liquidityPool.AmountOf(s.stakingDenom),
			DistHARP:                   harpPool.AmountOf(s.stakingDenom),
			DistModuleBalanceMain:      distrMAccBalance.AmountOf(s.mainDenom),
			DistModuleBalanceStaking:   distrMAccBalance.AmountOf(s.stakingDenom),
			DistBankBalanceMain:        distrBankBalance.AmountOf(s.mainDenom),
			DistBankBalanceStaking:     distrBankBalance.AmountOf(s.stakingDenom),
			DistLockedRatio:            s.QueryDistLockedRatio(),
			//
			SupplyTotalMain:    totalSupply.AmountOf(s.mainDenom),
			SupplyTotalStaking: totalSupply.AmountOf(s.stakingDenom),
			SupplyTotalLP:      totalSupply.AmountOf(s.lpDenom),
			//
			AccsBalanceMain:    accsTotalMain,
			AccsBalanceStaking: accsTotalStaking,
			//
			Counters: s.counter,
			//
			formatIntDecimals: func(value sdk.Int) string {
				return s.FormatIntDecimals(value, s.mainAmountDecimalsRatio)
			},
			formatDecDecimals: func(value sdk.Dec) string {
				return s.FormatDecDecimals(value, s.mainAmountDecimalsRatio)
			},
			formatCoinsDecimals: func(coins sdk.Coins) string {
				return s.FormatCoins(coins)
			},
		}

		// calculate statistics
		item.StatsBondedRatio = sdk.NewDecFromInt(item.StakingBonded.Add(item.StakingNotBonded)).Quo(sdk.NewDecFromInt(item.SupplyTotalStaking))
		item.StatsLPRatio = sdk.NewDecFromInt(item.StakingLPs).Quo(sdk.NewDecFromInt(item.SupplyTotalLP))
		reportItemIdx++

		for _, writer := range writers {
			writer.Write(item)
		}

		if debug {
			debugItem := BuildDebugReportItem(s)
			fmt.Println(debugItem.String())
		}

		return true, ""
	}

	return NewSimOperation(id, period, NewPeriodicNextExecFn(), handler)
}

type SimReportConsoleWriter struct {
	startedAt time.Time
}

func (w *SimReportConsoleWriter) Write(item SimReportItem) {
	reportingDur := time.Since(w.startedAt)

	str := strings.Builder{}

	str.WriteString(fmt.Sprintf("Report (%v):\n", reportingDur))
	str.WriteString(fmt.Sprintf("  BlockHeight:               %d\n", item.BlockHeight))
	str.WriteString(fmt.Sprintf("  BlockTime:                 %s\n", item.BlockTime.Format("02.01.2006T15:04:05")))
	str.WriteString(fmt.Sprintf("  SimDuration:               %v\n", FormatDuration(item.SimulationDur)))
	str.WriteString(fmt.Sprintf("   ValidatorsBonded:         %d\n", item.ValidatorsBonded))
	str.WriteString(fmt.Sprintf("   ValidatorsUnbonding:      %d\n", item.ValidatorsUnbonding))
	str.WriteString(fmt.Sprintf("   ValidatorsUnbonded:       %d\n", item.ValidatorsUnbonded))
	str.WriteString(fmt.Sprintf("    Staking: Bonded:         %s\n", item.formatIntDecimals(item.StakingBonded)))
	str.WriteString(fmt.Sprintf("    Staking: NotBonded:      %s\n", item.formatIntDecimals(item.StakingNotBonded)))
	str.WriteString(fmt.Sprintf("    Staking: Redelegations:  %d\n", item.RedelegationsInProcess))
	str.WriteString(fmt.Sprintf("     Mint: MinInflation:     %s\n", item.MintMinInflation))
	str.WriteString(fmt.Sprintf("     Mint: MaxInflation:     %s\n", item.MintMaxInflation))
	str.WriteString(fmt.Sprintf("     Mint: AnnualProvision:  %s\n", item.formatDecDecimals(item.MintAnnualProvisions)))
	str.WriteString(fmt.Sprintf("     Mint: BlocksPerYear:    %d\n", item.MintBlocksPerYear))
	str.WriteString(fmt.Sprintf("      Dist: FoundationPool:  %s\n", item.formatDecDecimals(item.DistFoundationPool)))
	str.WriteString(fmt.Sprintf("      Dist: PTreasuryPool:   %s\n", item.formatDecDecimals(item.DistPublicTreasuryPool)))
	str.WriteString(fmt.Sprintf("      Dist: LiquidityPPool:  %s\n", item.formatDecDecimals(item.DistLiquidityProvidersPool)))
	str.WriteString(fmt.Sprintf("      Dist: HARP:            %s\n", item.formatDecDecimals(item.DistHARP)))
	str.WriteString(fmt.Sprintf("      Dist: MAccBalance [m]  %s\n", item.formatIntDecimals(item.DistModuleBalanceMain)))
	str.WriteString(fmt.Sprintf("      Dist: MAccBalance [s]  %s\n", item.formatIntDecimals(item.DistModuleBalanceStaking)))
	str.WriteString(fmt.Sprintf("      Dist: BankBalance [m]  %s\n", item.formatIntDecimals(item.DistBankBalanceMain)))
	str.WriteString(fmt.Sprintf("      Dist: BankBalance [s]  %s\n", item.formatIntDecimals(item.DistBankBalanceStaking)))
	str.WriteString(fmt.Sprintf("      Dist: LockedRatio      %s\n", item.DistLockedRatio))
	str.WriteString(fmt.Sprintf("       Supply: TotalMain:    %s\n", item.formatIntDecimals(item.SupplyTotalMain)))
	str.WriteString(fmt.Sprintf("       Supply: TotalStaking: %s\n", item.formatIntDecimals(item.SupplyTotalStaking)))
	str.WriteString(fmt.Sprintf("       Supply: LPs:          %s\n", item.formatIntDecimals(item.SupplyTotalLP)))
	str.WriteString(fmt.Sprintf("  Stats: Bonded/TotalSupply [s]:  %s\n", item.StatsBondedRatio))
	str.WriteString(fmt.Sprintf("  Stats: Bonded/TotalSupply [LP]: %s\n", item.StatsLPRatio))
	str.WriteString(fmt.Sprintf("   Accounts: Balance [m]:         %s\n", item.formatIntDecimals(item.AccsBalanceMain)))
	str.WriteString(fmt.Sprintf("   Accounts: Balance [s]:         %s\n", item.formatIntDecimals(item.AccsBalanceStaking)))
	str.WriteString("  Counters:\n")
	str.WriteString("    Bonding:\n")
	str.WriteString(fmt.Sprintf("      Delegations:            %d\n", item.Counters.BDelegations))
	str.WriteString(fmt.Sprintf("      Redelegations:          %d\n", item.Counters.BRedelegations))
	str.WriteString(fmt.Sprintf("      Undelegations:          %d\n", item.Counters.BUndelegations))
	str.WriteString("    LP:\n")
	str.WriteString(fmt.Sprintf("      Delegations:            %d\n", item.Counters.LPDelegations))
	str.WriteString(fmt.Sprintf("      Redelegations:          %d\n", item.Counters.LPRedelegations))
	str.WriteString(fmt.Sprintf("      Undelegations:          %d\n", item.Counters.LPUndelegations))
	str.WriteString(fmt.Sprintf("    RewardWithdraws:          %d\n", item.Counters.RewardsWithdraws))
	str.WriteString(fmt.Sprintf("    RewardsCollected [m]:     %s\n", item.formatIntDecimals(item.Counters.RewardsCollectedMain)))
	str.WriteString(fmt.Sprintf("    RewardsCollected [s]:     %s\n", item.formatIntDecimals(item.Counters.RewardsCollectedStaking)))
	str.WriteString(fmt.Sprintf("    CommissionWithdraws:      %d\n", item.Counters.CommissionWithdraws))
	str.WriteString(fmt.Sprintf("    CommissionsCollected [m]: %s\n", item.formatIntDecimals(item.Counters.CommissionsCollectedMain)))
	str.WriteString(fmt.Sprintf("    CommissionsCollected [s]: %s\n", item.formatIntDecimals(item.Counters.CommissionsCollectedStaking)))
	str.WriteString(fmt.Sprintf("    Locked rewards:           %d\n", item.Counters.LockedRewards))

	fmt.Println(str.String())
}

func NewSimReportConsoleWriter() *SimReportConsoleWriter {
	w := &SimReportConsoleWriter{
		startedAt: time.Now(),
	}

	return w
}

// FormatDuration yet another duration formatter.
// 1.2.1 years -> 1 year, 2 months and 1 week
// 5.30 hours -> 5 hours and 30 minutes
func FormatDuration(dur time.Duration) string {
	dur = dur.Round(time.Minute)

	years := dur / Year
	dur -= years * Year
	months := dur / Month
	dur -= months * Month
	weeks := dur / Week
	dur -= weeks * Week
	days := dur / Day
	dur -= days * Day
	hours := dur / time.Hour
	dur -= hours * time.Hour
	mins := dur / time.Minute

	str := strings.Builder{}
	str.WriteString(fmt.Sprintf("%d years %d months %d weeks %d days ", years, months, weeks, days))
	str.WriteString(fmt.Sprintf("%d.%d hours", hours, mins))

	return str.String()
}
