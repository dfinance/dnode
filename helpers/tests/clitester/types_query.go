package clitester

import (
	"fmt"
	"strings"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
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

func (q *QueryRequest) SetNonQueryCmd(module string, args ...string) {
	q.cmd.AddArg("", module)

	for _, arg := range args {
		q.cmd.AddArg("", arg)
	}
}

func (q *QueryRequest) Execute() (combinedOutput string, retErr error) {
	code, stdout, stderr := q.cmd.Execute()
	combinedOutput = string(append(stdout, stderr...))

	if code != 0 {
		retErr = fmt.Errorf("%s: failed with code %d:\nstdout: %s\nstrerr: %s", q.String(), code, string(stdout), string(stderr))
		return
	}
	if len(stderr) > 0 {
		retErr = fmt.Errorf("%s: failed with non-empty stderr:\nstdout: %s\nstrerr: %s", q.String(), string(stdout), string(stderr))
		return
	}

	if q.resultObj != nil {
		if err := q.cdc.UnmarshalJSON(stdout, q.resultObj); err != nil {
			retErr = fmt.Errorf("%s: unmarshal query stdout: %s", q.String(), string(stdout))
			return
		}
	}

	return
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

func (q *QueryRequest) CheckFailedWithSDKError(err error) {
	sdkErr, ok := err.(*sdkErrors.Error)
	require.True(q.t, ok, "not a SDK error")

	code, stdout, stderr := q.cmd.Execute()
	require.NotEqual(q.t, 0, code, "%s: succeeded", q.String())
	trimmedStdout, trimmedStderr := trimCliOutput(stdout), trimCliOutput(stderr)

	qResponse := struct {
		Codespace string `json:"codespace"`
		Code      uint32 `json:"code"`
	}{"", 0}

	if err := q.cdc.UnmarshalJSON(trimmedStdout, &qResponse); err == nil {
		require.Equal(q.t, sdkErr.Codespace(), qResponse.Codespace, "%s: codespace", q.String())
		require.Equal(q.t, sdkErr.ABCICode(), qResponse.Code, "%s: code", q.String())
		return
	}

	if err := q.cdc.UnmarshalJSON(trimmedStderr, &qResponse); err == nil {
		require.Equal(q.t, sdkErr.Codespace(), qResponse.Codespace, "%s: codespace", q.String())
		require.Equal(q.t, sdkErr.ABCICode(), qResponse.Code, "%s: code", q.String())
		return
	}

	if strings.Contains(string(stdout), sdkErr.Error()) || strings.Contains(string(stderr), sdkErr.Error()) {
		return
	}

	q.t.Fatalf("%s: error %q can't be found in stdout/stderr: %s / %s", q.String(), err.Error(), string(stdout), string(stderr))
}

func (q *QueryRequest) CheckFailedWithErrorSubstring(subStr string) (output string) {
	code, stdout, stderr := q.cmd.Execute()
	require.NotEqual(q.t, 0, code, "%s: succeeded", q.String())

	stdoutStr, stderrStr := string(stdout), string(stderr)
	output = fmt.Sprintf("stdout: %s\nstderr: %s", stdoutStr, stderrStr)

	if subStr == "" {
		return
	}

	require.True(q.t,
		strings.Contains(stdoutStr, subStr) || strings.Contains(stderrStr, subStr),
		"%s: stdout/stderr doesn't contain %q sub string:\n%s",
		q.String(), subStr, output,
	)

	return
}

func (q *QueryRequest) String() string {
	return fmt.Sprintf("query %s", q.cmd.String())
}
