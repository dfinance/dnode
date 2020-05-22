package clitester

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"

	"github.com/dfinance/dnode/cmd/config"
)

type TxRequest struct {
	t              *testing.T
	cdc            *codec.Codec
	cmd            *CLICmd
	accPassphrase  string
	nodeRpcAddress string
	gas            uint64
}

func (r *TxRequest) SetCmd(module, fromAddress string, args ...string) {
	r.cmd.AddArg("", "tx")
	r.cmd.AddArg("", module)

	for _, arg := range args {
		r.cmd.AddArg("", arg)
	}

	if fromAddress != "" {
		r.cmd.AddArg("from", fromAddress)
	}
	r.cmd.AddArg("broadcast-mode", "block")
	r.cmd.AddArg("node", r.nodeRpcAddress)
	r.cmd.AddArg("fees", "1"+config.MainDenom)
	r.cmd.AddArg("gas", strconv.FormatUint(r.gas, 10))
	r.cmd.AddArg("", "--yes")
}

func (r *TxRequest) SetGas(amount uint64) *TxRequest {
	r.cmd.AddArg("gas", strconv.FormatUint(amount, 10))
	return r
}

func (r *TxRequest) DisableBroadcastMode() *TxRequest {
	r.cmd.RemoveArg("broadcast-mode")

	return r
}

func (r *TxRequest) ChangeCmdArg(oldArg, newArg string) *TxRequest {
	r.cmd.ChangeArg(oldArg, newArg)

	return r
}

func (r *TxRequest) RemoveCmdArg(arg string) *TxRequest {
	r.cmd.RemoveArg(arg)

	return r
}

func (r *TxRequest) Send() (retCode int, retStdout, retStderr []byte) {
	return r.cmd.Execute(r.accPassphrase, r.accPassphrase)
}

func (r *TxRequest) CheckSucceeded() {
	code, stdout, stderr := r.Send()

	require.Equal(r.t, 0, code, "%s: failed with code %d:\nstdout: %s\nstrerr: %s", r.String(), code, string(stdout), string(stderr))
	require.Len(r.t, stderr, 0, "%s: failed with non-empty stderr:\nstdout: %s\nstrerr: %s", r.String(), string(stdout), string(stderr))

	if len(stdout) > 0 {
		txResponse := sdk.TxResponse{}
		require.NoError(r.t, r.cdc.UnmarshalJSON(stdout, &txResponse), "%s: unmarshal", r.String())
		require.Equal(r.t, "", txResponse.Codespace, "%s: codespace: %s", r.String(), string(stdout))
		require.Equal(r.t, uint32(0), txResponse.Code, "%s: code: %s", r.String(), string(stdout))
	}
}

func (r *TxRequest) CheckFailedWithSDKError(err error) {
	sdkErr, ok := err.(*sdkErrors.Error)
	require.True(r.t, ok, "not a SDK error")

	code, stdout, stderr := r.Send()
	require.NotEqual(r.t, 0, code, "%s: succeeded", r.String())
	stdout, stderr = trimCliOutput(stdout), trimCliOutput(stderr)

	txResponse := sdk.TxResponse{}
	stdoutErr := r.cdc.UnmarshalJSON(stdout, &txResponse)
	stderrErr := r.cdc.UnmarshalJSON(stderr, &txResponse)
	if stdoutErr != nil && stderrErr != nil {
		r.t.Fatalf("%s: unmarshal stdout/stderr: %s / %s", r.String(), string(stdout), string(stderr))
	}

	require.Equal(r.t, sdkErr.Codespace(), txResponse.Codespace, "%s: codespace", r.String())
	require.Equal(r.t, sdkErr.ABCICode(), txResponse.Code, "%s: code", r.String())
}

func (r *TxRequest) CheckFailedWithErrorSubstring(subStr string) (output string) {
	code, stdout, stderr := r.Send()
	require.NotEqual(r.t, 0, code, "%s: succeeded", r.String())

	stdoutStr, stderrErr := string(stdout), string(stderr)
	output = fmt.Sprintf("stdout: %s\nstderr: %s", stdoutStr, stderrErr)

	if subStr == "" {
		return
	}

	if strings.Contains(stdoutStr, subStr) || strings.Contains(stderrErr, subStr) {
		return
	}
	r.t.Fatalf("%s: stdout/stderr doesn't contain %q sub string: %s", r.String(), subStr, output)

	return
}

func (r *TxRequest) String() string {
	return fmt.Sprintf("tx %s", r.cmd.String())
}
