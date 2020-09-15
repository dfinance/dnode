package simulator

import (
	"encoding/csv"
	"github.com/stretchr/testify/require"
	"os"
	"strconv"
	"testing"
	"time"
)

type SimReportCSVWriter struct {
	startedAt time.Time
	writer    *csv.Writer
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
}

func NewSimReportCSVWriter(t *testing.T, fileName string) (*SimReportCSVWriter, CSVWriterClose) {
	file, err := os.Create(fileName)
	require.Nil(t, err)

	closeFn := func() {
		file.Close()
	}

	writer := csv.NewWriter(file)
	err = writer.Write(Headers)
	require.Nil(t, err)

	return &SimReportCSVWriter{
		startedAt: time.Now(),
		writer:    writer,
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
		item.StatsBondedRation.String(),
		strconv.FormatInt(item.Counters.Delegations, 10),
		strconv.FormatInt(item.Counters.Redelegations, 10),
		strconv.FormatInt(item.Counters.Undelegations, 10),
		strconv.FormatInt(item.Counters.Rewards, 10),
	}

	w.writer.Write(data)
}
