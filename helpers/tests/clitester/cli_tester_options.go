package clitester

import (
	"fmt"
)

type CLITesterOption func(ct *CLITester) error

func VMConnectionSettings(minBackoffMs, maxBackoffMs, maxAttempts int) CLITesterOption {
	return func(ct *CLITester) error {
		ct.vmComMinBackoffMs = minBackoffMs
		ct.vmComMaxBackoffMs = maxBackoffMs
		ct.vmComMaxAttempts = maxAttempts

		return nil
	}
}

func VMCommunicationBaseAddressNet(baseAddr string) CLITesterOption {
	return func(ct *CLITester) error {
		ct.vmBaseAddress = baseAddr
		ct.vmConnectAddress = fmt.Sprintf("%s:%s", ct.vmBaseAddress, ct.VmConnectPort)
		ct.vmListenAddress = fmt.Sprintf("%s:%s", ct.vmBaseAddress, ct.VmListenPort)

		return nil
	}
}

func VMCommunicationBaseAddressUDS(listenFileName, vmFileName string) CLITesterOption {
	return func(ct *CLITester) error {
		ct.vmBaseAddress = "unix://" + ct.UDSDir
		ct.vmConnectAddress = fmt.Sprintf("%s/%s", ct.vmBaseAddress, vmFileName)
		ct.vmListenAddress = fmt.Sprintf("%s/%s", ct.vmBaseAddress, listenFileName)

		return nil
	}
}
