package clitester

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/tests"
	"github.com/stretchr/testify/require"
)

type CLICmd struct {
	sync.Mutex
	t            *testing.T
	base         string
	args         []string
	inputs       string
	proc         *tests.Process
	logs         []string
	pipeLoggerWG sync.WaitGroup
}

func (c *CLICmd) AddArg(flagName, flagValue string) *CLICmd {
	if flagName != "" {
		c.args = append(c.args, fmt.Sprintf("--%s=%s", flagName, flagValue))
	} else {
		c.args = append(c.args, flagValue)
	}

	return c
}

func (c *CLICmd) ChangeArg(oldArg, newArg string) *CLICmd {
	for i := 0; i < len(c.args); i++ {
		if c.args[i] == oldArg {
			c.args[i] = newArg
			break
		}

		if ("--" + c.args[i]) == oldArg {
			c.args[i] = "--" + newArg
			break
		}
	}

	return c
}

func (c *CLICmd) RemoveArg(arg string) *CLICmd {
	argAlt := "--" + arg
	for i := 0; i < len(c.args); i++ {
		if strings.HasPrefix(c.args[i], arg) || strings.HasPrefix(c.args[i], argAlt) {
			c.args = append(c.args[:i], c.args[i+1:]...)
			break
		}
	}

	return c
}

func (c *CLICmd) String() string {
	return fmt.Sprintf("cmd %q with args [%s] and inputs [%s]", c.base, strings.Join(c.args, " "), c.inputs)
}

func (c *CLICmd) Execute(stdinInput ...string) (retCode int, retStdout, retStderr []byte) {
	c.inputs = strings.Join(stdinInput, ", ")

	proc, err := tests.StartProcess("", c.base, c.args)
	require.NoError(c.t, err, "cmd %q: StartProcess", c.String())

	for _, input := range stdinInput {
		_, err := proc.StdinPipe.Write([]byte(input + "\n"))
		require.NoError(c.t, err, "%s: %q StdinPipe.Write", c.String(), input)
	}

	stdout, stderr, err := proc.ReadAll()
	require.NoError(c.t, err, "%s: reading stdout, stderr", c.String())

	proc.Wait()
	retCode, retStdout, retStderr = proc.ExitState.ExitCode(), stdout, stderr

	return
}

func (c *CLICmd) Start(t *testing.T, printLogs bool) {
	proc, err := tests.CreateProcess("", c.base, c.args)
	require.NoError(c.t, err, "cmd %q: CreateProcess", c.String())

	pipeLogger := func(pipeName string, pipe io.ReadCloser, wg *sync.WaitGroup) {
		defer wg.Done()

		buf := bufio.NewReader(pipe)
		for {
			line, _, err := buf.ReadLine()
			if err != nil {
				if printLogs {
					c.t.Logf("%q %s: reading daemon pipe: %v", c.base, pipeName, err)
				}
				return
			}
			logMsg := fmt.Sprintf("%s->%s: %s", c.base, pipeName, line)

			if printLogs {
				t.Log(logMsg)
			}

			c.Lock()
			c.logs = append(c.logs, logMsg)
			c.Unlock()
		}
	}

	c.pipeLoggerWG = sync.WaitGroup{}
	c.pipeLoggerWG.Add(2)

	go pipeLogger("stdout", proc.StdoutPipe, &c.pipeLoggerWG)
	go pipeLogger("stderr", proc.StderrPipe, &c.pipeLoggerWG)

	require.NoError(c.t, proc.Cmd.Start(), "cmd %q: Start", c.String())
	c.proc = proc
}

func (c *CLICmd) Stop() {
	require.NotNil(c.t, c.proc, "proc")
	require.NoError(c.t, c.proc.Stop(true), "proc.Stop")
	c.proc = nil
}

func (c *CLICmd) WaitForStop(timeout time.Duration) (exitCode *int) {
	require.NotNil(c.t, c.proc, "proc")

	timeoutCh := time.NewTimer(timeout).C
	stopCh := make(chan bool)

	go func() {
		c.proc.Wait()
		state := c.proc.ExitState
		if state != nil {
			code := state.ExitCode()
			exitCode = &code
		}

		c.pipeLoggerWG.Wait()
		close(stopCh)
	}()

	select {
	case <-timeoutCh:
	case <-stopCh:
	}

	return
}

func (c *CLICmd) CheckSuccessfulExecute(resultObj interface{}, stdinInput ...string) {
	code, stdout, stderr := c.Execute(stdinInput...)
	require.Equal(c.t, 0, code, "%s: stderr: %s", c.String(), string(stderr))

	if resultObj != nil {
		if err := json.Unmarshal(stdout, resultObj); err == nil {
			return
		}
		if err := json.Unmarshal(stderr, resultObj); err == nil {
			return
		}

		c.t.Fatalf("%s: stdout/stderr unmarshal to object type %t", c.String(), resultObj)
	}
}

func (c *CLICmd) LogsContain(subStr string) bool {
	c.Lock()
	defer c.Unlock()

	for _, l := range c.logs {
		if strings.Contains(l, subStr) {
			return true
		}
	}

	return false
}
