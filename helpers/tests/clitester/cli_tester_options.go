package clitester

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type AccountOption struct {
	Name     string
	Balances []StringPair
}

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

func LogLevel(logLevel string) CLITesterOption {
	return func(ct *CLITester) error {
		ct.daemonLogLvl = logLevel

		return nil
	}
}

func DefaultConsensusTimings() CLITesterOption {
	return func(ct *CLITester) error {
		ct.defaultConsensusTimeouts = true

		return nil
	}
}

func Accounts(accOpts ...AccountOption) CLITesterOption {
	return func(ct *CLITester) error {
		for _, opt := range accOpts {

			account := &CLIAccount{
				Name:  opt.Name,
				Coins: make(map[string]sdk.Coin, len(opt.Balances)),
			}
			for _, balance := range opt.Balances {
				amount, ok := sdk.NewIntFromString(balance.Value)
				if !ok {
					return fmt.Errorf("sdk.NewIntFromString for %q: failed", balance.Value)
				}

				account.Coins[balance.Key] = sdk.NewCoin(balance.Key, amount)
			}

			ct.Accounts[opt.Name] = account
		}

		return nil
	}
}
