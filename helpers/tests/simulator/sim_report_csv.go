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
	"Staking: Bonded",
	"Staking: Redelegations",
	"Staking: NotBonded",
	"Mint: MinInflation",
	"Mint: MaxInflation",
	"Mint: AnnualProvision",
	"Mint: BlocksPerYear",
	"Dist: FoundationPool",
	"Dist: PTreasuryPool",
	"Dist: LiquidityPPool",
	"Dist: HARP",
	"Supply: Total",
	"Stats: Bonded/TotalSupply",
	"Counters: Delegations:",
	"Counters: Redelegations:",
	"Counters: Undelegations:",
	"Counters: Rewards:",
	"Counters: RewardsCollected:",
	"Counters: Commissions:",
	"Counters: CommissionsCollected:",
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
		item.StakingBonded.String(),
		strconv.Itoa(item.RedelegationsInProcess),
		item.StakingNotBonded.String(),
		item.MintMinInflation.String(),
		item.MintMaxInflation.String(),
		item.MintAnnualProvisions.String(),
		strconv.FormatUint(item.MintBlocksPerYear, 10),
		item.DistFoundationPool.String(),
		item.DistPublicTreasuryPool.String(),
		item.DistLiquidityProvidersPool.String(),
		item.DistHARP.String(),
		item.SupplyTotal.String(),
		item.StatsBondedRatio.String(),
		strconv.FormatInt(item.Counters.Delegations, 10),
		strconv.FormatInt(item.Counters.Redelegations, 10),
		strconv.FormatInt(item.Counters.Undelegations, 10),
		strconv.FormatInt(item.Counters.Rewards, 10),
		item.Counters.RewardsCollected.String(),
		strconv.FormatInt(item.Counters.Commissions, 10),
		item.Counters.CommissionsCollected.String(),
	}

	_ = w.writer.Write(data)
}
