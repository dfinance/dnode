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
	EnvDvmIntegUse            = "DN_DVM_INTEG_TESTS_USE"
	EnvDvmIntegDockerRegistry = "DN_DVM_INTEG_TESTS_DOCKER_REGISTRY"
	EnvDvmIntegDockerTag      = "DN_DVM_INTEG_TESTS_DOCKER_TAG"
	EnvDvmIntegBinaryPath     = "DN_DVM_INTEG_TESTS_BINARY_PATH"
	//
	EnvDvmIntegUseDocker = "docker"
	EnvDvmIntegUseBinary = "binary"
	//
	TestErrFmt = "Launching DVM over %s with %s transport"
)

func LaunchDVMWithNetTransport(t *testing.T, connectPort, dsServerPort string, printLogs bool, args ...string) (stopFunc func()) {
	transportLabel := "Net"

	if ok, registry, tag, errMsg := dvmDockerLaunchEnvParams(transportLabel); ok {
		container, err := docker.NewDVMWithNetTransport(registry, tag, connectPort, dsServerPort, printLogs, args...)
		require.NoError(t, err, errMsg)

		return func() {
			if err := container.Stop(); err != nil {
				t.Logf("stopping container: %v", err)
			}
		}
	}

	if ok, path, errMsg := dvmBinaryLaunchEnvParams(transportLabel); ok {
		cmd, err := binary.NewDVMWithNetTransport(path, connectPort, dsServerPort, printLogs, args...)
		require.NoError(t, err, errMsg)

		return func() {
			if err := cmd.Stop(); err != nil {
				t.Logf("stopping binary: %v", err)
			}
		}
	}

	t.Fatalf("Docker / Binary DVM launch option not specified: %s", os.Getenv(EnvDvmIntegUse))

	return nil
}

func LaunchDVMWithUDSTransport(t *testing.T, socketsDir, connectSocketName, dsSocketName string, printLogs bool, args ...string) (stopFunc func()) {
	transportLabel := "UDS"

	if ok, registry, tag, errMsg := dvmDockerLaunchEnvParams(transportLabel); ok {
		container, err := docker.NewDVMWithUDSTransport(registry, tag, socketsDir, connectSocketName, dsSocketName, printLogs, args...)
		require.NoError(t, err, errMsg)

		return func() {
			if err := container.Stop(); err != nil {
				t.Logf("stopping container: %v", err)
			}
		}
	}

	if ok, cmdPath, errMsg := dvmBinaryLaunchEnvParams(transportLabel); ok {
		cmd, err := binary.NewDVMWithUDSTransport(cmdPath, socketsDir, connectSocketName, dsSocketName, printLogs, args...)
		require.NoError(t, err, errMsg)

		return func() {
			if err := cmd.Stop(); err != nil {
				t.Logf("stopping binary: %v", err)
			}
		}
	}

	t.Fatalf("Docker / Binary DVM launch option not specified: %s", os.Getenv(EnvDvmIntegUse))

	return nil
}

func dvmDockerLaunchEnvParams(transportLabel string) (enabled bool, registry, tag, errMsg string) {
	if os.Getenv(EnvDvmIntegUse) != EnvDvmIntegUseDocker {
		return
	}
	enabled = true
	registry = os.Getenv(EnvDvmIntegDockerRegistry)
	tag = os.Getenv(EnvDvmIntegDockerTag)
	errMsg = fmt.Sprintf(TestErrFmt, "Docker", transportLabel)

	return
}

func dvmBinaryLaunchEnvParams(transportLabel string) (enabled bool, path, errMsg string) {
	if os.Getenv(EnvDvmIntegUse) != EnvDvmIntegUseBinary {
		return
	}
	enabled = true
	path = os.Getenv(EnvDvmIntegBinaryPath)
	errMsg = fmt.Sprintf(TestErrFmt, "UDS", transportLabel)

	return
}
