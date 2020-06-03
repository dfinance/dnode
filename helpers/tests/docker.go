package tests

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/server"
	docker "github.com/fsouza/go-dockerclient"
)

type DockerContainerOption func(*DockerContainer) error

type DockerContainer struct {
	dClient    *docker.Client
	dContainer *docker.Container
	dOptions   docker.CreateContainerOptions
}

func NewDockerContainer(options ...DockerContainerOption) (*DockerContainer, error) {
	c := DockerContainer{}

	c.dOptions = docker.CreateContainerOptions{
		Config:     &docker.Config{},
		HostConfig: &docker.HostConfig{},
	}

	for _, options := range options {
		if err := options(&c); err != nil {
			return nil, err
		}
	}

	return &c, nil
}

func WithCreds(registry, name, tag string) DockerContainerOption {
	return func(c *DockerContainer) error {
		c.dOptions.Config.Image = fmt.Sprintf("%s/%s:%s", registry, name, tag)
		return nil
	}
}

func WithCmdArgs(cmdArgs []string) DockerContainerOption {
	return func(c *DockerContainer) error {
		c.dOptions.Config.Cmd = cmdArgs
		return nil
	}
}

func WithVolume(hostPath, containerPath string) DockerContainerOption {
	return func(c *DockerContainer) error {
		c.dOptions.HostConfig.VolumeDriver = "bind"
		c.dOptions.HostConfig.Binds = append(
			c.dOptions.HostConfig.Binds,
			fmt.Sprintf("%s:%s", hostPath, containerPath),
		)

		return nil
	}
}

func WithTcpPorts(tcpPorts []string) DockerContainerOption {
	return func(c *DockerContainer) error {
		ports := make(map[docker.Port]struct{}, len(tcpPorts))
		portBindings := make(map[docker.Port][]docker.PortBinding, len(tcpPorts))
		for _, p := range tcpPorts {
			dPort := docker.Port(fmt.Sprintf("%s/tcp", p))

			ports[dPort] = struct{}{}
			portBindings[dPort] = []docker.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: dPort.Port(),
				},
			}
		}

		c.dOptions.Config.ExposedPorts = ports
		c.dOptions.HostConfig.PortBindings = portBindings

		return nil
	}
}

func WithHostNetwork() DockerContainerOption {
	return func(c *DockerContainer) error {
		_, mode, err := HostMachineDockerUrl()
		if err != nil {
			return err
		}

		c.dOptions.HostConfig.NetworkMode = mode

		return nil
	}
}

func WithUser() DockerContainerOption {
	return func(c *DockerContainer) error {
		userUid, userGid := os.Getuid(), os.Getgid()
		if userUid < 0 {
			return fmt.Errorf("invalid user UID: %d", userUid)
		}
		if userGid < 0 {
			return fmt.Errorf("invalid user GID: %d", userGid)
		}

		c.dOptions.Config.User = fmt.Sprintf("%d:%d", userUid, userGid)

		return nil
	}
}

func (c *DockerContainer) String() string {
	return c.dOptions.Config.Image
}

func (c *DockerContainer) Start(startTimeout time.Duration) error {
	if c.dClient != nil {
		return fmt.Errorf("%q: already started", c.String())
	}

	client, err := docker.NewClientFromEnv()
	if err != nil {
		return fmt.Errorf("%q: connecting to docker: %w", c.String(), err)
	}

	container, err := client.CreateContainer(c.dOptions)
	if err != nil {
		return fmt.Errorf("%q: creating container: %w", c.String(), err)
	}

	if err := client.StartContainer(container.ID, nil); err != nil {
		return fmt.Errorf("%q: starting container: %w", c.String(), err)
	}

	// wait for container to start
	timeoutCh := time.NewTimer(startTimeout).C
	for {
		inspectContainer, err := client.InspectContainerWithOptions(docker.InspectContainerOptions{ID: container.ID})
		if err != nil {
			return fmt.Errorf("%q: wait for container to start: %w", c.String(), err)
		}
		if inspectContainer.State.Running {
			break
		}

		select {
		case <-timeoutCh:
			return fmt.Errorf("%q: wait for container to start: timeout reached (%v)", c.String(), startTimeout)
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}

	// wait for all TCP port to be reachable
	portReports := make(map[docker.Port]string)
	for p := range c.dOptions.Config.ExposedPorts {
		portReports[p] = "not checked"
	}
	for {
		cnt := len(portReports)
		for p, status := range portReports {
			if status == "OK" {
				cnt--
				continue
			}

			if err := PingTcpAddress("127.0.0.1:" + p.Port()); err != nil {
				portReports[p] = err.Error()
			} else {
				portReports[p] = "OK"
				cnt--
				continue
			}

			select {
			case <-timeoutCh:
				reports := make([]string, 0, len(portReports))
				for p, status := range portReports {
					reports = append(reports, fmt.Sprintf("%s: %s", p.Port(), status))
				}

				return fmt.Errorf(
					"%q: wait for container TCP ports to be rechable: timeout reached (%v): %s",
					c.String(),
					startTimeout,
					strings.Join(reports, ", "),
				)
			default:
				time.Sleep(100 * time.Millisecond)
			}
		}
		if cnt == 0 {
			break
		}
	}

	c.dClient = client
	c.dContainer = container

	return nil
}

func (c *DockerContainer) Stop() error {
	if c.dClient == nil {
		return fmt.Errorf("%q: not started", c.String())
	}

	err := c.dClient.RemoveContainer(docker.RemoveContainerOptions{
		ID:    c.dContainer.ID,
		Force: true,
	})
	if err != nil {
		return fmt.Errorf("%q: removing container: %w", c.String(), err)
	}

	return nil
}

func NewVMCompilerContainerWithNetTransport(dsServerPort string) (retContainer *DockerContainer, retPort string, retErr error) {
	_, port, err := server.FreeTCPAddr()
	if err != nil {
		retErr = fmt.Errorf("FreeTCPAddr (VMCompiler): %w", err)
		return
	}
	retPort = port

	tag := os.Getenv("TAG")
	if tag == "" {
		tag = "master"
	}

	registry := os.Getenv("REGISTRY")
	if registry == "" {
		retErr = fmt.Errorf("REGISTRY env var: not found")
		return
	}

	hostUrl, _, _ := HostMachineDockerUrl()
	dsServerAddress := fmt.Sprintf("%s:%s", hostUrl, dsServerPort)
	cmdArgs := []string{"./compiler", "http://0.0.0.0:" + port, dsServerAddress}

	retContainer, retErr = NewDockerContainer(
		WithCreds(registry, "dfinance/dvm", tag),
		WithCmdArgs(cmdArgs),
		WithTcpPorts([]string{port}),
		WithHostNetwork(),
	)

	return
}

func NewVMRuntimeContainerWithNetTransport(connectPort, dsServerPort string) (retContainer *DockerContainer, retErr error) {
	tag := os.Getenv("TAG")
	if tag == "" {
		tag = "master"
	}

	registry := os.Getenv("REGISTRY")
	if registry == "" {
		retErr = fmt.Errorf("REGISTRY env var: not found")
		return
	}

	hostUrl, _, _ := HostMachineDockerUrl()
	dsServerAddress := fmt.Sprintf("%s:%s", hostUrl, dsServerPort)
	cmdArgs := []string{"./dvm", "http://0.0.0.0:" + connectPort, dsServerAddress}

	retContainer, retErr = NewDockerContainer(
		WithCreds(registry, "dfinance/dvm", tag),
		WithCmdArgs(cmdArgs),
		WithTcpPorts([]string{connectPort}),
		WithHostNetwork(),
	)

	return
}

func NewVMCompilerContainerWithUDSTransport(volumePath, dsFileName, vmFileName string) (retContainer *DockerContainer, retErr error) {
	const defVolumePath = "/tmp/dn-uds"

	tag := os.Getenv("TAG")
	if tag == "" {
		tag = "master"
	}

	registry := os.Getenv("REGISTRY")
	if registry == "" {
		retErr = fmt.Errorf("REGISTRY env var: not found")
		return
	}

	dsFilePath := path.Join(defVolumePath, dsFileName)
	vmFilePath := path.Join(defVolumePath, vmFileName)

	// one '/' is omitted on purpose
	cmdArgs := []string{"./compiler", "-v", "ipc:/" + vmFilePath, "ipc:/" + dsFilePath}

	retContainer, retErr = NewDockerContainer(
		WithCreds(registry, "dfinance/dvm", tag),
		WithCmdArgs(cmdArgs),
		WithVolume(volumePath, defVolumePath),
		WithUser(),
	)

	return
}

func NewVMRuntimeContainerWithUDSTransport(volumePath, dsFileName, vmFileName string) (retContainer *DockerContainer, retErr error) {
	const defVolumePath = "/tmp/dn-uds"

	tag := os.Getenv("TAG")
	if tag == "" {
		tag = "master"
	}

	registry := os.Getenv("REGISTRY")
	if registry == "" {
		retErr = fmt.Errorf("REGISTRY env var: not found")
		return
	}

	dsFilePath := path.Join(defVolumePath, dsFileName)
	vmFilePath := path.Join(defVolumePath, vmFileName)

	// one '/' is omitted on purpose
	cmdArgs := []string{"./dvm", "-v", "ipc:/" + vmFilePath, "ipc:/" + dsFilePath}

	retContainer, retErr = NewDockerContainer(
		WithCreds(registry, "dfinance/dvm", tag),
		WithCmdArgs(cmdArgs),
		WithVolume(volumePath, defVolumePath),
		WithUser(),
	)

	return
}

func HostMachineDockerUrl() (hostUrl, hostNetworkMode string, err error) {
	switch runtime.GOOS {
	case "darwin", "windows":
		hostUrl, hostNetworkMode = "http://host.docker.internal", ""
	case "linux":
		hostUrl, hostNetworkMode = "http://localhost", "host"
	default:
		err = fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}

	return
}
