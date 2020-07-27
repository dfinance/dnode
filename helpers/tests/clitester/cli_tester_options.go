package clitester

import (
	"fmt"
	"path"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type AccountOption struct {
	Name     string
	Balances []StringPair
}

type CLITesterOption func(ct *CLITester) error

func RootDirectoryOption(rootDir, socketsDirName string) CLITesterOption {
	return func(ct *CLITester) error {
		dncliDir := path.Join(rootDir, "dncli")
		udsDir := path.Join(rootDir, socketsDirName)

		ct.Dirs = DirConfig{
			RootDir:  rootDir,
			DncliDir: dncliDir,
			UDSDir:   udsDir,
		}

		return nil
	}
}

func NodeIDOption(chainID, monikerID string) CLITesterOption {
	return func(ct *CLITester) error {
		ct.IDs = NodeIdConfig{
			ChainID:   chainID,
			MonikerID: monikerID,
		}

		return nil
	}
}

func BinaryPathsOptions(dnodePath, dncliPath string) CLITesterOption {
	return func(ct *CLITester) error {
		ct.BinaryPath = BinaryPathConfig{
			wbd:   dnodePath,
			wbcli: dncliPath,
		}

		return nil
	}
}

func VMCommunicationOption(minBackoffMs, maxBackoffMs, maxAttempts int) CLITesterOption {
	return func(ct *CLITester) error {
		ct.VMCommunication.MinBackoffMs = minBackoffMs
		ct.VMCommunication.MaxBackoffMs = maxBackoffMs
		ct.VMCommunication.MaxAttempts = maxAttempts

		return nil
	}
}

func VMCommunicationBaseAddressNetOption(baseAddr string) CLITesterOption {
	return func(ct *CLITester) error {
		ct.VMConnection.BaseAddress = baseAddr
		ct.VMConnection.ConnectAddress = fmt.Sprintf("%s:%s", ct.VMConnection.BaseAddress, ct.VMConnection.ConnectPort)
		ct.VMConnection.ListenAddress = fmt.Sprintf("%s:%s", ct.VMConnection.BaseAddress, ct.VMConnection.ListenPort)
		ct.VMConnection.CompilerAddress = ct.VMConnection.ConnectAddress

		return nil
	}
}

func VMCommunicationBaseAddressUDSOption(listenFileName, vmFileName string) CLITesterOption {
	return func(ct *CLITester) error {
		ct.VMConnection.BaseAddress = "unix://" + ct.Dirs.UDSDir
		ct.VMConnection.ConnectAddress = fmt.Sprintf("%s/%s", ct.VMConnection.BaseAddress, vmFileName)
		ct.VMConnection.ListenAddress = fmt.Sprintf("%s/%s", ct.VMConnection.BaseAddress, listenFileName)
		ct.VMConnection.CompilerAddress = ct.VMConnection.ConnectAddress

		return nil
	}
}

func DefaultConsensusTimingsOption() CLITesterOption {
	return func(ct *CLITester) error {
		ct.ConsensusTimings.UseDefaults = true

		return nil
	}
}

func ConsensusTimingsOption(propose, proposeDelta, preVote, preVoteDelta, preCommit, preCommitDelta, commit string) CLITesterOption {
	return func(ct *CLITester) error {
		ct.ConsensusTimings.UseDefaults = false
		ct.ConsensusTimings.TimeoutPropose = propose
		ct.ConsensusTimings.TimeoutProposeDelta = proposeDelta
		ct.ConsensusTimings.TimeoutPreVote = preVote
		ct.ConsensusTimings.TimeoutPreVoteDelta = preVoteDelta
		ct.ConsensusTimings.TimeoutPreCommit = preCommit
		ct.ConsensusTimings.TimeoutPreCommitDelta = preCommitDelta
		ct.ConsensusTimings.TimeoutCommit = commit

		return nil
	}
}

func MempoolOption(size, cacheSize, maxTxBytes, maxTxsBytes int64) CLITesterOption {
	return func(ct *CLITester) error {
		ct.MempoolConfig.UseDefault = false
		ct.MempoolConfig.Size = size
		ct.MempoolConfig.CacheSize = cacheSize
		ct.MempoolConfig.MaxTxBytes = maxTxBytes
		ct.MempoolConfig.MaxTxsBytes = maxTxsBytes

		return nil
	}
}

func DaemonLogLevelOption(logLevel string) CLITesterOption {
	return func(ct *CLITester) error {
		ct.daemonLogLvl = logLevel

		return nil
	}
}

func AccountsOption(accOpts ...AccountOption) CLITesterOption {
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
