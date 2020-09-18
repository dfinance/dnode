// +build simulator

package simulator

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/dfinance/dnode/cmd/config"
)

const (
	reportFileName = "/tmp/report.csv"
)

func TestSimInflation(t *testing.T) {
	expSimDur := 3 * 30 * 24 * time.Hour

	// genesis accounts balance (1M xfi)
	genAmount, ok := sdk.NewIntFromString("1000000000000000000000000")
	require.True(t, ok)
	genCoin := sdk.NewCoin(config.MainDenom, genAmount)
	delegationCoin := sdk.NewCoin(config.MainDenom, genAmount.QuoRaw(10))

	// custom distribution params
	treasuryCapacity, ok := sdk.NewIntFromString("250000000000000000000000")
	require.True(t, ok)
	distParams := distribution.DefaultParams()
	distParams.PublicTreasuryPoolCapacity = treasuryCapacity

	reportWriter, writerClose := NewSimReportCSVWriter(t, reportFileName)
	defer writerClose()

	// create simulator
	s := NewSimulator(t, NewDefferOps(),
		GenerateWalletAccountsOption(5, 3, sdk.NewCoins(genCoin)),
		LogOption(log.AllowInfoWith("module", "x/mint")),
		LogOption(log.AllowInfoWith("module", "x/distribution")),
		DistributionParamsOption(distParams),
		OperationsOption(
			NewReportOp(1*time.Hour, reportWriter),
			NewReportOp(1*time.Hour, NewSimReportConsoleWriter()),
			NewCreateValidatorOp(30*time.Minute),
			NewDelegateOp(60*time.Minute, delegationCoin),
			NewRedelegateOp(120*time.Minute),
			NewUndelegateOp(180*time.Minute),
			NewTakeReward(30*time.Minute),
			NewTakeCommission(60*time.Minute),
		),
	)
	s.Start()

	// work loop
	_, simDur := s.SimulatedDur()
	for simDur < expSimDur {
		s.Next()

		_, simDur = s.SimulatedDur()
	}
}
