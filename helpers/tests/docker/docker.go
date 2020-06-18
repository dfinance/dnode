package docker

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"runtime"
	"strings"
	"time"

	docker "github.com/fsouza/go-dockerclient"

	testUtils "github.com/dfinance/dnode/helpers/tests/utils"
)

const (
	DvmDockerStartTimeout = 5 * time.Second
)

type DockerContainerOption func(*DockerContainer) error

type DockerContainer struct {
	dClient    *docker.Client
	dContainer *docker.Container
	dOptions   docker.CreateContainerOptions
	printLogs  bool
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

func WithConsoleLogs(enabled bool) DockerContainerOption {
	return func(c *DockerContainer) error {
		c.printLogs = enabled
		return nil
	}
}

func (c *DockerContainer) String() string {
	return "container " + c.dOptions.Config.Image
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

	if c.printLogs {
		stdoutBuf, stderrBuf := bytes.Buffer{}, bytes.Buffer{}
		opts := docker.LogsOptions{
			Container:    container.ID,
			OutputStream: &stdoutBuf,
			ErrorStream:  &stderrBuf,
			Follow:       true,
			Stdout:       true,
			Stderr:       true,
			Tail: "0",
			Timestamps:   false,
		}

		//if err := client.Logs(opts); err != nil {
		//	return fmt.Errorf("%q: setting log options: %w", c.String(), err)
		//}

		streamPrinter := func(streamName string, buf *bytes.Buffer) {
			for {
				line, err := buf.ReadString('\n')
				if err != nil && err != io.EOF {
					fmt.Printf("%s: broken %s stream: %v\n", c.String(), streamName, err)
					return
				}

				line = strings.TrimSpace(line)
				if line != "" {
					fmt.Printf("%s: %s stream: %s\n", c.String(), streamName, line)
				}
			}
		}

		client.Logs(opts)
		go streamPrinter("stdout", &stdoutBuf)
		go streamPrinter("stderr", &stderrBuf)
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

			if err := testUtils.PingTcpAddress("127.0.0.1:"+p.Port(), 500*time.Millisecond); err != nil {
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

func NewDVMWithNetTransport(registry, tag, connectPort, dsServerPort string, printLogs bool, args ...string) (*DockerContainer, error) {
	if registry == "" || tag == "" {
		return nil, fmt.Errorf("registry / tag: not specified")
	}

	hostUrl, _, _ := HostMachineDockerUrl()
	dsServerAddress := fmt.Sprintf("%s:%s", hostUrl, dsServerPort)
	cmdArgs := []string{"./dvm", "http://0.0.0.0:" + connectPort, dsServerAddress}
	if len(args) > 0 {
		cmdArgs = append(cmdArgs, strings.Join(args, " "))
	}

	container, err := NewDockerContainer(
		WithCreds(registry, "dfinance/dvm", tag),
		WithCmdArgs(cmdArgs),
		WithTcpPorts([]string{connectPort}),
		WithHostNetwork(),
		WithConsoleLogs(printLogs),
	)
	if err != nil {
		return nil, fmt.Errorf("creating DVM container over Net: %v", err)
	}

	if err := container.Start(DvmDockerStartTimeout); err != nil {
		return nil, fmt.Errorf("starting DVM container over Net: %v", err)
	}

	return container, nil
}

func NewDVMWithUDSTransport(registry, tag, volumePath, vmFileName, dsFileName string, printLogs bool, args ...string) (*DockerContainer, error) {
	const defVolumePath = "/tmp/dn-uds"

	if registry == "" || tag == "" {
		return nil, fmt.Errorf("registry / tag: not specified")
	}

	vmFilePath := path.Join(defVolumePath, vmFileName)
	dsFilePath := path.Join(defVolumePath, dsFileName)

	// one '/' is omitted on purpose
	cmdArgs := []string{"./dvm", "ipc:/" + vmFilePath, "ipc:/" + dsFilePath}
	if len(args) > 0 {
		cmdArgs = append(cmdArgs, strings.Join(args, " "))
	}

	container, err := NewDockerContainer(
		WithCreds(registry, "dfinance/dvm", tag),
		WithCmdArgs(cmdArgs),
		WithVolume(volumePath, defVolumePath),
		WithUser(),
		WithConsoleLogs(printLogs),
	)
	if err != nil {
		return nil, fmt.Errorf("creating DVM container over UDS: %v", err)
	}

	if err := container.Start(DvmDockerStartTimeout); err != nil {
		return nil, fmt.Errorf("starting DVM container over UDS: %v", err)
	}

	if err := testUtils.WaitForFileExists(path.Join(volumePath, vmFileName), DvmDockerStartTimeout); err != nil {
		return nil, fmt.Errorf("creating DVM container over UDS: %v", err)
	}

	return container, nil
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
