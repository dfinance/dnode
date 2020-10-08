package simulator

import (
	"encoding/csv"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

type SimReportCSVWriter struct {
	writer *csv.Writer
}

type CSVWriterClose func()

var Headers = []string{
	"BlockHeight",
	"BlockTime",
	"SimDuration",
	"Validators: Bonded",
	"Validators: Unbonding",
	"Validators: Unbonded",
	"Staking: Bonded",
	"Staking: NotBonded",
	"Staking: LPs",
	"Staking: ActiveRedelegations",
	"Mint: MinInflation",
	"Mint: MaxInflation",
	"Mint: AnnualProvision",
	"Mint: BlocksPerYear",
	"Dist: FoundationPool",
	"Dist: PTreasuryPool",
	"Dist: LiquidityPPool",
	"Dist: HARP",
	"Dist: MAccBalance [main]",
	"Dist: MAccBalance [staking]",
	"Dist: BankBalance [main]",
	"Dist: BankBalance [staking]",
	"Dist: LockedRatio",
	"Supply: Total [main]",
	"Supply: Total [staking]",
	"Supply: Total [LP]",
	"Stats: Staked/TotalSupply [staking]",
	"Stats: Staked/TotalSupply [LPs]",
	"Accounts: TotalBalance [main]",
	"Accounts: TotalBalance [staking]",
	"Counters: Bonding: Delegations",
	"Counters: Bonding: Redelegations",
	"Counters: Bonding: Undelegations",
	"Counters: LP: Delegations",
	"Counters: LP: Redelegations",
	"Counters: LP: Undelegations",
	"Counters: RewardWithdraws",
	"Counters: RewardsCollected [main]",
	"Counters: RewardsCollected [staking]",
	"Counters: CommissionWithdraws",
	"Counters: CommissionsCollected [main]",
	"Counters: CommissionsCollected [staking]",
	"Counters: LockedRewards",
}

func NewSimReportCSVWriter(t *testing.T, filePath string) (*SimReportCSVWriter, CSVWriterClose) {
	file, err := os.Create(filePath)
	require.Nil(t, err)

	closeFn := func() {
		file.Close()
	}

	writer := csv.NewWriter(file)
	err = writer.Write(Headers)
	require.Nil(t, err)

	return &SimReportCSVWriter{
		writer: writer,
	}, closeFn
}

func (w *SimReportCSVWriter) Write(item SimReportItem) {
	defer w.writer.Flush()

	data := []string{
		strconv.FormatInt(item.BlockHeight, 10),
		item.BlockTime.Format("02.01.2006T15:04:05"),
		FormatDuration(item.SimulationDur),
		// validators
		strconv.Itoa(item.ValidatorsBonded),
		strconv.Itoa(item.ValidatorsUnbonding),
		strconv.Itoa(item.ValidatorsUnbonded),
		// staking
		item.formatIntDecimals(item.StakingBonded),
		item.formatIntDecimals(item.StakingNotBonded),
		item.formatIntDecimals(item.StakingLPs),
		strconv.Itoa(item.RedelegationsInProcess),
		// mint
		item.MintMinInflation.String(),
		item.MintMaxInflation.String(),
		item.formatDecDecimals(item.MintAnnualProvisions),
		strconv.FormatUint(item.MintBlocksPerYear, 10),
		// distribution
		item.formatDecDecimals(item.DistFoundationPool),
		item.formatDecDecimals(item.DistPublicTreasuryPool),
		item.formatDecDecimals(item.DistLiquidityProvidersPool),
		item.formatDecDecimals(item.DistHARP),
		item.formatIntDecimals(item.DistModuleBalanceMain),
		item.formatIntDecimals(item.DistModuleBalanceStaking),
		item.formatIntDecimals(item.DistBankBalanceMain),
		item.formatIntDecimals(item.DistBankBalanceStaking),
		item.DistLockedRatio.String(),
		// supply
		item.formatIntDecimals(item.SupplyTotalMain),
		item.formatIntDecimals(item.SupplyTotalStaking),
		item.formatIntDecimals(item.SupplyTotalLP),
		// stats
		item.StatsBondedRatio.String(),
		item.StatsLPRatio.String(),
		// accounts
		item.formatIntDecimals(item.AccsBalanceMain),
		item.formatIntDecimals(item.AccsBalanceStaking),
		// counters
		// - bonding
		strconv.FormatInt(item.Counters.BDelegations, 10),
		strconv.FormatInt(item.Counters.BRedelegations, 10),
		strconv.FormatInt(item.Counters.BUndelegations, 10),
		// - LP
		strconv.FormatInt(item.Counters.LPDelegations, 10),
		strconv.FormatInt(item.Counters.LPRedelegations, 10),
		strconv.FormatInt(item.Counters.LPUndelegations, 10),
		// - rewards
		strconv.FormatInt(item.Counters.RewardsWithdraws, 10),
		item.formatIntDecimals(item.Counters.RewardsCollectedMain),
		item.formatIntDecimals(item.Counters.RewardsCollectedStaking),
		// - commissions
		strconv.FormatInt(item.Counters.CommissionWithdraws, 10),
		item.formatIntDecimals(item.Counters.CommissionsCollectedMain),
		item.formatIntDecimals(item.Counters.CommissionsCollectedStaking),
		// - locking
		strconv.FormatInt(item.Counters.LockedRewards, 10),
	}

	_ = w.writer.Write(data)
}
