// +build simulator

package simulator

import (
	"fmt"
	"io/ioutil"
	"net/http"
	_ "net/http/pprof"
	"path"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/dfinance/dnode/cmd/config"
)

func TestSimInflation(t *testing.T) {
	go http.ListenAndServe(":8090", nil)

	expSimDur := 24 * 30 * 24 * time.Hour

	// create a tmp directory
	workingDir, err := ioutil.TempDir("/tmp", fmt.Sprintf("dnode-simulator-%s-", t.Name()))
	require.NoError(t, err)

	// genesis accounts balance (1M xfi, 1M sxfi)
	genAmount, ok := sdk.NewIntFromString("1000000000000000000000000")
	require.True(t, ok)
	genCoins := sdk.NewCoins(
		sdk.NewCoin(config.MainDenom, genAmount),
		sdk.NewCoin(config.StakingDenom, genAmount),
	)

	// custom distribution params
	treasuryCapacity, ok := sdk.NewIntFromString("250000000000000000000000")
	require.True(t, ok)
	distParams := distribution.DefaultParams()
	distParams.PublicTreasuryPoolCapacity = treasuryCapacity

	// custom staking params
	stakingParams := staking.DefaultParams()
	stakingParams.UnbondingTime = 24 * time.Hour
	stakingParams.MaxValidators = 100

	// CSV report writer
	reportWriter, writerClose := NewSimReportCSVWriter(t, path.Join(workingDir, "report.csv"))
	defer writerClose()

	// create simulator
	s := NewSimulator(t, workingDir, NewDefferOps(),
		//InMemoryDBOption(),
		BlockTimeOption(60*time.Second, 65*time.Second),
		GenerateWalletAccountsOption(500, 3, 150, genCoins),
		LogOption(log.AllowInfoWith("module", "x/staking")),
		LogOption(log.AllowInfoWith("module", "x/mint")),
		LogOption(log.AllowInfoWith("module", "x/distribution")),
		LogOption(log.AllowInfoWith("module", "x/slashing")),
		LogOption(log.AllowInfoWith("module", "x/evidence")),
		StakingParamsOption(stakingParams),
		DistributionParamsOption(distParams),
		InvariantCheckPeriodOption(1000),
		OperationsOption(
			NewSimInvariantsOp(1*time.Hour),
			NewForceUpdateOp(8*time.Hour),
			//
			NewReportOp(24*time.Hour, false, NewSimReportConsoleWriter(), reportWriter),
			//
			NewCreateValidatorOp(2*24*time.Hour),
			NewDelegateOp(8*time.Hour, sdk.NewDecWithPrec(40, 2)),    // 40 %
			NewRedelegateOp(20*time.Hour, sdk.NewDecWithPrec(20, 2)), // 20 %
			NewUndelegateOp(48*time.Hour, sdk.NewDecWithPrec(25, 2)), // 25 %
			//
			NewGetDelRewardOp(120*time.Hour),
			NewGetValRewardOp(72*time.Hour),
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
