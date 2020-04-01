package clitester

import (
	"fmt"
	"strings"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

type QueryRequest struct {
	t              *testing.T
	cdc            *codec.Codec
	cmd            *CLICmd
	nodeRpcAddress string
	resultObj      interface{}
}

func (q *QueryRequest) ChangeCmdArg(oldArg, newArg string) {
	q.cmd.ChangeArg(oldArg, newArg)
}

func (q *QueryRequest) RemoveCmdArg(arg string) {
	q.cmd.RemoveArg(arg)
}

func (q *QueryRequest) SetCmd(module string, args ...string) {
	q.cmd.AddArg("", "query")
	q.cmd.AddArg("", module)

	for _, arg := range args {
		q.cmd.AddArg("", arg)
	}

	q.cmd.AddArg("node", q.nodeRpcAddress)
}

func (q *QueryRequest) CheckSucceeded() {
	code, stdout, stderr := q.cmd.Execute()

	require.Equal(q.t, 0, code, "%s: failed with code %d:\nstdout: %s\nstrerr: %s", q.String(), code, string(stdout), string(stderr))
	require.Len(q.t, stderr, 0, "%s: failed with non-empty stderr:\nstdout: %s\nstrerr: %s", q.String(), string(stdout), string(stderr))

	if q.resultObj != nil {
		err := q.cdc.UnmarshalJSON(stdout, q.resultObj)
		require.NoError(q.t, err, "%s: unmarshal query stdout: %s", q.String(), string(stdout))
	}
}

func (q *QueryRequest) CheckFailedWithSDKError(sdkErr sdk.Error) {
	code, stdout, stderr := q.cmd.Execute()
	require.NotEqual(q.t, 0, code, "%s: succeeded", q.String())
	stdout, stderr = trimCliOutput(stdout), trimCliOutput(stderr)

	qResponse := struct {
		Codespace sdk.CodespaceType `json:"codespace"`
		Code      sdk.CodeType      `json:"code"`
	}{sdk.CodespaceType(""), sdk.CodeType(0)}
	stdoutErr := q.cdc.UnmarshalJSON(stdout, &qResponse)
	stderrErr := q.cdc.UnmarshalJSON(stderr, &qResponse)
	if stdoutErr != nil && stderrErr != nil {
		q.t.Fatalf("%s: unmarshal stdout/stderr: %s / %s", q.String(), string(stdout), string(stderr))
	}

	require.Equal(q.t, sdkErr.Codespace(), qResponse.Codespace, "%s: codespace", q.String())
	require.Equal(q.t, sdkErr.Code(), qResponse.Code, "%s: code", q.String())
}

func (q *QueryRequest) CheckFailedWithErrorSubstring(subStr string) (output string) {
	code, stdout, stderr := q.cmd.Execute()
	require.NotEqual(q.t, 0, code, "%s: succeeded", q.String())

	stdoutStr, stderrErr := string(stdout), string(stderr)
	output = fmt.Sprintf("stdout: %s\nstderr: %s", stdoutStr, stderrErr)

	if subStr == "" {
		return
	}

	if strings.Contains(stdoutStr, subStr) || strings.Contains(stderrErr, subStr) {
		return
	}
	q.t.Fatalf("%s: stdout/stderr doesn't contain %q sub string", q.String(), subStr)

	return
}

func (q *QueryRequest) String() string {
	return fmt.Sprintf("query %s", q.cmd.String())
}
