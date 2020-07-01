package clitester

import (
	"os"
	"path"
	"strings"

	sdkKeys "github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"github.com/dfinance/dnode/cmd/config"
	dnConfig "github.com/dfinance/dnode/cmd/config"
)

func (ct *CLITester) initChain() {
	// init chain
	cmd := ct.newWbdCmd().AddArg("", "init").AddArg("", ct.IDs.MonikerID).AddArg("chain-id", ct.IDs.ChainID)
	cmd.CheckSuccessfulExecute(nil)

	// configure dncli
	{
		cmd := ct.newWbcliCmd().
			AddArg("", "config").
			AddArg("", "keyring-backend").
			AddArg("", string(ct.keyringBackend))
		cmd.CheckSuccessfulExecute(nil)
	}

	// adjust Tendermint config (make blocks generation faster)
	{
		if !ct.ConsensusTimings.UseDefaults {
			cfgMtx.Lock()
			defer cfgMtx.Unlock()

			cfgFile := path.Join(ct.Dirs.RootDir, "config", "config.toml")
			_, err := os.Stat(cfgFile)
			require.NoError(ct.t, err, "reading config.toml file")
			viper.SetConfigFile(cfgFile)
			require.NoError(ct.t, viper.ReadInConfig())

			viper.Set("consensus.timeout_propose", ct.ConsensusTimings.TimeoutPropose)
			viper.Set("consensus.timeout_propose_delta", ct.ConsensusTimings.TimeoutProposeDelta)
			viper.Set("consensus.timeout_prevote", ct.ConsensusTimings.TimeoutPreVote)
			viper.Set("consensus.timeout_prevote_delta", ct.ConsensusTimings.TimeoutPreVoteDelta)
			viper.Set("consensus.timeout_precommit", ct.ConsensusTimings.TimeoutPreCommit)
			viper.Set("consensus.timeout_precommit_delta", ct.ConsensusTimings.TimeoutPreCommitDelta)
			viper.Set("consensus.timeout_commit", ct.ConsensusTimings.TimeoutCommit)

			require.NoError(ct.t, viper.WriteConfig(), "saving config.toml file")
		}
	}

	// configure accounts
	{
		poaValidatorIdx := 0
		for accName, accValue := range ct.Accounts {
			// create key
			{
				cmd := ct.newWbcliCmd().
					AddArg("", "keys").
					AddArg("", "add").
					AddArg("", accName)
				output := sdkKeys.KeyOutput{}

				cmd.CheckSuccessfulExecute(&output, ct.AccountPassphrase, ct.AccountPassphrase)
				accValue.Name = output.Name
				accValue.Address = output.Address
				accValue.PubKey = output.PubKey
				accValue.Mnemonic = output.Mnemonic
			}

			// get armored private key
			{
				cmd := ct.newWbcliCmd().
					AddArg("", "keys").
					AddArg("", "export").
					AddArg("", accName)

				output := cmd.CheckSuccessfulExecute(nil, ct.AccountPassphrase, ct.AccountPassphrase, ct.AccountPassphrase, ct.AccountPassphrase)
				require.NoError(ct.t, ct.keyBase.ImportPrivKey(accName, output, ct.AccountPassphrase), "account %q: keyBase.ImportPrivKey", accName)
			}

			// genesis account
			{
				cmd := ct.newWbdCmd().
					AddArg("", "add-genesis-account").
					AddArg("", accValue.Address)
				if accValue.IsModuleAcc {
					cmd.AddArg("module-name", accName)
				}

				require.NotEmpty(ct.t, accValue.Coins, "account %q: no coins", accName)
				var coinsArg []string
				for _, coin := range accValue.Coins {
					coinsArg = append(coinsArg, coin.String())
				}
				cmd.AddArg("", strings.Join(coinsArg, ","))

				cmd.CheckSuccessfulExecute(nil, ct.AccountPassphrase)
			}

			// POA validator
			if accValue.IsPOAValidator {
				require.True(ct.t, poaValidatorIdx < len(EthAddresses), "add more predefined ethAddresses")
				accValue.EthAddress = EthAddresses[poaValidatorIdx]

				cmd := ct.newWbdCmd().
					AddArg("", "add-genesis-poa-validator").
					AddArg("", accValue.Address).
					AddArg("", accValue.EthAddress)
				cmd.CheckSuccessfulExecute(nil, ct.AccountPassphrase)

				poaValidatorIdx++
			}

			// Oracle nominee
			if accValue.IsOracleNominee {
				cmd := ct.newWbdCmd().
					AddArg("", "add-oracle-nominees-gen").
					AddArg("", accValue.Address)

				cmd.CheckSuccessfulExecute(nil, ct.AccountPassphrase)
			}
		}
	}

	// validator genTX
	{
		stakingCoin := ct.Accounts["pos"].Coins[config.MainDenom]
		stakingCoin.Amount = stakingCoin.Amount.QuoRaw(4)

		cmd := ct.newWbdCmd().
			AddArg("", "gentx").
			AddArg("home-client", ct.Dirs.DncliDir).
			AddArg("name", "pos").
			AddArg("amount", stakingCoin.String()).
			AddArg("keyring-backend", string(ct.keyringBackend))

		cmd.CheckSuccessfulExecute(nil, ct.AccountPassphrase, ct.AccountPassphrase, ct.AccountPassphrase)
	}

	// VM default write sets
	{
		defWriteSetsPath := os.ExpandEnv(DefVmWriteSetsPath)

		cmd := ct.newWbdCmd().
			AddArg("", "read-genesis-write-set").
			AddArg("", defWriteSetsPath).
			AddArg("home", ct.Dirs.RootDir)

		cmd.CheckSuccessfulExecute(nil)
	}

	// add Oracle assets
	{
		oracles := make([]string, 0)
		oracles = append(oracles, ct.Accounts["oracle1"].Address)
		oracles = append(oracles, ct.Accounts["oracle2"].Address)

		cmd := ct.newWbdCmd().
			AddArg("", "add-oracle-asset-gen").
			AddArg("", ct.DefAssetCode).
			AddArg("", strings.Join(oracles, ","))

		cmd.CheckSuccessfulExecute(nil)
	}

	// adjust governance genesis
	{
		appState := ct.GenesisState()

		govGenesis := gov.GenesisState{}
		require.NoError(ct.t, ct.Cdc.UnmarshalJSON(appState[gov.ModuleName], &govGenesis), "unmarshal gov genesisState")

		govGenesis.VotingParams.VotingPeriod = ct.GovernanceConfig.MinVotingDur
		govGenesisRaw, err := ct.Cdc.MarshalJSON(govGenesis)
		require.NoError(ct.t, err, "marshal gov genesisState")
		appState[gov.ModuleName] = govGenesisRaw

		ct.updateGenesisState(appState)
	}

	// collect genTXs
	{
		cmd := ct.newWbdCmd().AddArg("", "collect-gentxs")
		cmd.CheckSuccessfulExecute(nil)
	}

	// validate genesis
	{
		cmd := ct.newWbdCmd().AddArg("", "validate-genesis")
		cmd.CheckSuccessfulExecute(nil)
	}

	// prepare VM config
	{
		vmConfig := dnConfig.DefaultVMConfig()
		vmConfig.Address, vmConfig.DataListen = ct.VMConnection.ConnectAddress, ct.VMConnection.ListenAddress
		vmConfig.InitialBackoff = ct.VMCommunication.MinBackoffMs
		vmConfig.MaxBackoff = ct.VMCommunication.MaxBackoffMs
		vmConfig.MaxAttempts = ct.VMCommunication.MaxAttempts
		dnConfig.WriteVMConfig(ct.Dirs.RootDir, vmConfig)
	}
}
