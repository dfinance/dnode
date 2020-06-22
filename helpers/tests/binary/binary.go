package binary

import (
	"bufio"
	"fmt"
	"io"
	"path"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/tests"

	testUtils "github.com/dfinance/dnode/helpers/tests/utils"
)

const (
	DvmBinaryStartTimeout = 2 * time.Second
)

type BinaryCmdOption func(*BinaryCmd) error

type BinaryCmd struct {
	cmd          string
	proc         *tests.Process
	args         []string
	printLogs    bool
}

func (c *BinaryCmd) String() string {
	return fmt.Sprintf("binary %s %s", c.cmd, strings.Join(c.args, " "))
}

func NewBinaryCmd(cmd string, options ...BinaryCmdOption) (*BinaryCmd, error) {
	c := &BinaryCmd{
		cmd:     cmd,
	}

	for _, option := range options {
		if err := option(c); err != nil {
			return nil, fmt.Errorf("%s: option apply failed: %w", c.String(), err)
		}
	}

	return c, nil
}

func WithArgs(args ...string) BinaryCmdOption {
	return func(c *BinaryCmd) error {
		c.args = args
		return nil
	}
}

func WithConsoleLogs(enabled bool) BinaryCmdOption {
	return func(c *BinaryCmd) error {
		c.printLogs = enabled
		return nil
	}
}

func (c *BinaryCmd) Start() error {
	if c.proc != nil {
		return fmt.Errorf("%s: process already started", c.String())
	}

	proc, err := tests.CreateProcess("", c.cmd, c.args)
	if err != nil {
		return fmt.Errorf("%s: creating process: %w", c.String(), err)
	}

	if c.printLogs {
		pipeLogger := func(msgFmtPrefix string, pipe io.ReadCloser) {
			msgFmt := msgFmtPrefix + "%s: %s" + testUtils.FmtColorEndLine

			buf := bufio.NewReader(pipe)
			for {
				line, _, err := buf.ReadLine()
				if err != nil {
					if err == io.EOF {
						return
					}

					fmt.Printf("%s: broken pipe: %v", c.String(), err.Error())
					return
				}

				fmt.Printf(msgFmt, c.String(), line)
			}
		}

		go pipeLogger(testUtils.FmtInfColorPrefix, proc.StdoutPipe)
		go pipeLogger(testUtils.FmtWrnColorPrefix, proc.StderrPipe)
	}

	if err := proc.Cmd.Start(); err != nil {
		return fmt.Errorf("%s: starting process: %w", c.String(), err)
	}
	c.proc = proc

	return nil
}

func (c *BinaryCmd) Stop() error {
	if c.proc == nil {
		return nil
	}

	if err := c.proc.Stop(true); err != nil {
		return fmt.Errorf("%s: stop failed: %w", c.String(), err)
	}

	return nil
}

func NewDVMWithNetTransport(basePath, connectPort, dsServerPort string, printLogs bool, args ...string) (*BinaryCmd, error) {
	cmdArgs := []string{
		"http://127.0.0.1:" + connectPort,
		"http://127.0.0.1:" + dsServerPort,
	}
	cmdArgs = append(cmdArgs, args...)

	c, err := NewBinaryCmd(path.Join(basePath, "dvm"), WithArgs(cmdArgs...), WithConsoleLogs(printLogs))
	if err != nil {
		return nil, err
	}

	if err := c.Start(); err != nil {
		return nil, err
	}
	time.Sleep(DvmBinaryStartTimeout)

	return c, nil
}

func NewDVMWithUDSTransport(basePath, socketsDir, connectSocketName, dsSocketName string, printLogs bool, args ...string) (*BinaryCmd, error) {
	cmdArgs := []string{
		"ipc:/" + path.Join(socketsDir, connectSocketName),
		"ipc:/" + path.Join(socketsDir, dsSocketName),
	}
	cmdArgs = append(cmdArgs, args...)

	c, err := NewBinaryCmd(path.Join(basePath, "dvm"), WithArgs(cmdArgs...), WithConsoleLogs(printLogs))
	if err != nil {
		return nil, err
	}

	if err := c.Start(); err != nil {
		return nil, err
	}

	if err := testUtils.WaitForFileExists(path.Join(socketsDir, connectSocketName), DvmBinaryStartTimeout); err != nil {
		return nil, fmt.Errorf("%s: waiting for UDS server to start-up: %v", c.Start(), err)
	}

	return c, nil
}
