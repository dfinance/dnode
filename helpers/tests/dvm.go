package tests

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dfinance/dnode/helpers/tests/binary"
	"github.com/dfinance/dnode/helpers/tests/docker"
)

const (
	EnvDvmIntegDockerUse      = "DN_DVM_INTEG_TESTS_DOCKER_USE"
	EnvDvmIntegDockerRegistry = "DN_DVM_INTEG_TESTS_DOCKER_REGISTRY"
	EnvDvmIntegDockerTag      = "DN_DVM_INTEG_TESTS_DOCKER_TAG"
	EnvDvmIntegBinaryUse      = "DN_DVM_INTEG_TESTS_BINARY_USE"
	EnvDvmIntegBinaryPath     = "DN_DVM_INTEG_TESTS_BINARY_PATH"
	//
	TestErrFmt = "Launching DVM over %s with %s transport"
)

func LaunchDVMWithNetTransport(t *testing.T, connectPort, dsServerPort string, printLogs bool, args ...string) (stopFunc func()) {
	transportLabel := "Net"

	if ok, registry, tag, errMsg := dvmDockerLaunchEnvParams(transportLabel); ok {
		container, err := docker.NewDVMWithNetTransport(registry, tag, connectPort, dsServerPort, printLogs, args...)
		require.NoError(t, err, errMsg)

		return func() {
			container.Stop()
		}
	}

	if ok, path, errMsg := dvmBinaryLaunchEnvParams(transportLabel); ok {
		cmd, err := binary.NewDVMWithNetTransport(path, connectPort, dsServerPort, printLogs, args...)
		require.NoError(t, err, errMsg)

		return func() {
			cmd.Stop()
		}
	}

	t.Fatalf("Docker / Binary DVM launch option not specified")

	return nil
}

func LaunchDVMWithUDSTransport(t *testing.T, socketsDir, connectSocketName, dsSocketName string, printLogs bool, args ...string) (stopFunc func()) {
	transportLabel := "UDS"

	if ok, registry, tag, errMsg := dvmDockerLaunchEnvParams(transportLabel); ok {
		container, err := docker.NewDVMWithUDSTransport(registry, tag, socketsDir, connectSocketName, dsSocketName, printLogs, args...)
		require.NoError(t, err, errMsg)

		return func() {
			container.Stop()
		}
	}

	if ok, cmdPath, errMsg := dvmBinaryLaunchEnvParams(transportLabel); ok {
		cmd, err := binary.NewDVMWithUDSTransport(cmdPath, socketsDir, connectSocketName, dsSocketName, printLogs, args...)
		require.NoError(t, err, errMsg)

		return func() {
			cmd.Stop()
		}
	}

	t.Fatalf("Docker / Binary DVM launch option not specified")

	return nil
}

func dvmDockerLaunchEnvParams(transportLabel string) (enabled bool, registry, tag, errMsg string) {
	_, enabled = os.LookupEnv(EnvDvmIntegDockerUse)
	registry = os.Getenv(EnvDvmIntegDockerRegistry)
	tag = os.Getenv(EnvDvmIntegDockerTag)
	errMsg = fmt.Sprintf(TestErrFmt, "Docker", transportLabel)

	return
}

func dvmBinaryLaunchEnvParams(transportLabel string) (enabled bool, path, errMsg string) {
	_, enabled = os.LookupEnv(EnvDvmIntegBinaryUse)
	path = os.Getenv(EnvDvmIntegBinaryPath)
	errMsg = fmt.Sprintf(TestErrFmt, "UDS", transportLabel)

	return
}
