package clitester

import (
	"context"
	"os"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	tmClient "github.com/tendermint/tendermint/rpc/client"
	coreTypes "github.com/tendermint/tendermint/rpc/core/types"
)

func (ct *CLITester) CreateWSConnection(printLogs bool, subscriber, query string, chCap int) (func(), <-chan coreTypes.ResultEvent) {
	logger := log.NewNopLogger()
	if printLogs {
		logger = log.NewTMLogger(log.NewSyncWriter(os.Stdout))
	}

	client, err := tmClient.NewHTTP(ct.rpcAddress, "/websocket")
	require.NoError(ct.t, err, "creating WebSocket client")
	client.SetLogger(logger)
	require.NoError(ct.t, client.Start(), "starting WebSocket client")

	out, err := client.Subscribe(context.Background(), subscriber, query, chCap)
	require.NoError(ct.t, err, "WebSocket subscribe")

	stopFunc := func() {
		client.Stop()
	}

	return stopFunc, out
}
