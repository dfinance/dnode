package tests

import (
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/server"
	docker "github.com/fsouza/go-dockerclient"
)

type DockerContainerOption func(*DockerContainer)

type DockerContainer struct {
	dClient    *docker.Client
	dContainer *docker.Container
	dOptions   docker.CreateContainerOptions
}

func NewDockerContainer(options ...DockerContainerOption) *DockerContainer {
	c := DockerContainer{}

	c.dOptions = docker.CreateContainerOptions{
		Config:     &docker.Config{},
		HostConfig: &docker.HostConfig{},
	}

	for _, options := range options {
		options(&c)
	}

	return &c
}

func WithCreds(registry, name, tag string) DockerContainerOption {
	return func(c *DockerContainer) {
		c.dOptions.Config.Image = fmt.Sprintf("%s/%s:%s", registry, name, tag)
	}
}

func WithCmdArgs(cmdArgs []string) DockerContainerOption {
	return func(c *DockerContainer) {
		c.dOptions.Config.Cmd = cmdArgs
	}
}

func WithTcpPorts(tcpPorts []string) DockerContainerOption {
	return func(c *DockerContainer) {
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
	portReports := make(map[docker.Port]string, 0)
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

func NewVMCompilerContainer(dsServerPort string) (retContainer *DockerContainer, retPort string, retErr error) {
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

	dsServerAddress := fmt.Sprintf("http://%s:%s", "host.docker.internal", dsServerPort)
	cmdArgs := []string{"./compiler", "0.0.0.0:" + port, dsServerAddress}

	retContainer = NewDockerContainer(
		WithCreds(registry, "dfinance/dvm", tag),
		WithCmdArgs(cmdArgs),
		WithTcpPorts([]string{port}))

	return
}

func PingTcpAddress(address string) error {
	conn, err := net.DialTimeout("tcp", address, 500*time.Millisecond)
	if err != nil {
		return err
	}
	defer conn.Close()

	return nil
}
