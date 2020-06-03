package clitester

import (
	"context"
	"fmt"
	"os"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	tmClient "github.com/tendermint/tendermint/rpc/client"
	coreTypes "github.com/tendermint/tendermint/rpc/core/types"
)

func (ct *CLITester) CheckWSSubscribed(printLogs bool, subscriber, query string, chCap int) (func(), <-chan coreTypes.ResultEvent) {
	stopFunc, ch, err := ct.CreateWSConnection(printLogs, subscriber, query, chCap)
	require.NoError(ct.t, err, "WebSocket for %q query", query)

	return stopFunc, ch
}

func (ct *CLITester) CreateWSConnection(printLogs bool, subscriber, query string, chCap int) (retStopFunc func(), retCh <-chan coreTypes.ResultEvent, retErr error) {
	logger := log.NewNopLogger()
	if printLogs {
		logger = log.NewTMLogger(log.NewSyncWriter(os.Stdout))
	}

	client, err := tmClient.NewHTTP(ct.NodePorts.RPCAddress, "/websocket")
	if err != nil {
		retErr = fmt.Errorf("creating WebSocket client: %w", err)
		return
	}
	client.SetLogger(logger)

	if err := client.Start(); err != nil {
		retErr = fmt.Errorf("starting WebSocket client: %w", err)
		return
	}

	ch, err := client.Subscribe(context.Background(), subscriber, query, chCap)
	if err != nil {
		retErr = fmt.Errorf("WebSocket subscribe: %w", err)
		return
	}

	retStopFunc = func() {
		if err := client.Stop(); err != nil {
			logger.Error(fmt.Sprintf("stopping WSClient: %v", err))
		}
	}
	retCh = ch

	return
}
