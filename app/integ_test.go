// +build integ

package app

import (
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/dfinance/dnode/helpers/tests"
	cliTester "github.com/dfinance/dnode/helpers/tests/clitester"
)

func Test_ConsensusFailure(t *testing.T) {
	const script = `
import 0x0.Account;
import 0x0.Coins;
main(recipient: address, amount: u128, denom: bytearray) {
    let coin: Coins.Coin;
    coin = Account.withdraw_from_sender(move(amount), move(denom));
    Account.deposit(move(recipient), move(coin));
    return;
}
`

	t.Parallel()
	ct := cliTester.New(t, false)
	defer ct.Close()

	//ct.SetVMCompilerAddress("rpc.demo.wings.toys:50053")

	compilerContainer, compilerPort, err := tests.NewVMCompilerContainer(ct.VmListenPort)
	require.NoError(t, err, "compiler container creation")

	require.NoError(t, compilerContainer.Start(5*time.Second), "compiler container creation")
	defer compilerContainer.Stop()

	ct.SetVMCompilerAddress("127.0.0.1:" + compilerPort)

	senderAddr := ct.Accounts["validator1"].Address
	mvirPath := path.Join(ct.RootDir, "script.mvir")
	compiledPath := path.Join(ct.RootDir, "script.json")

	// Create .mvir script file
	mvirFile, err := os.Create(mvirPath)
	require.NoError(t, err, "creating script file")
	_, err = mvirFile.WriteString(script)
	require.NoError(t, err, "write script file")
	require.NoError(t, mvirFile.Close(), "close script file")

	// Compile .mvir script file
	ct.QueryVmCompileScript(mvirPath, compiledPath, senderAddr).CheckSucceeded()

	// Execute .json script file
	// Should panic as there is no local VM running
	ct.TxVmExecuteScript(senderAddr, compiledPath, senderAddr, "100", "dfi").DisableBroadcastMode().CheckSucceeded()

	// Check CONSENSUS FAILURE did occur
	{
		consensusFailure := false
		for i := 0; i < 10; i++ {
			if ct.DaemonLogsContain("CONSENSUS FAILURE") {
				consensusFailure = true
				break
			}
			time.Sleep(500 * time.Millisecond)
		}
		require.True(t, consensusFailure, "CONSENSUS FAILURE not occurred")
	}

	// Check restarted application panics
	{
		ct.RestartDaemon(false, false)

		retCode, daemonLogs := ct.CheckDaemonStopped(2 * time.Second)

		require.NotZero(t, retCode, "daemon exitCode")
		require.Contains(t, strings.Join(daemonLogs, ","), "panic", "daemon didn't panic")
	}
}
