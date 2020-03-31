package main

import (
	"os"
	"path/filepath"
	"time"

	sdkSrv "github.com/cosmos/cosmos-sdk/server"
	sdkSrvCfg "github.com/cosmos/cosmos-sdk/server/config"
	sdkVersion "github.com/cosmos/cosmos-sdk/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	tmCmd "github.com/tendermint/tendermint/cmd/tendermint/commands"
	tmCfg "github.com/tendermint/tendermint/config"
	tmCli "github.com/tendermint/tendermint/libs/cli"
	tmFlags "github.com/tendermint/tendermint/libs/cli/flags"
	tmLog "github.com/tendermint/tendermint/libs/log"
)

// Copy from: "github.com/cosmos/cosmos-sdk/server"
// Justification: original version overwrites Context.Logger and custom logger can't be used
// Original description:
//   PersistentPreRunEFn returns a PersistentPreRunE function for cobra
//   that initailizes the passed in context with a properly configured
//   logger and config object.
func PersistentPreRunEFn(context *sdkSrv.Context) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if cmd.Name() == sdkVersion.Cmd.Name() {
			return nil
		}
		config, err := interceptLoadConfig()
		if err != nil {
			return err
		}
		logger := context.Logger
		logger, err = tmFlags.ParseLogLevel(config.LogLevel, logger, tmCfg.DefaultLogLevel())
		if err != nil {
			return err
		}
		if viper.GetBool(tmCli.TraceFlag) {
			logger = tmLog.NewTracingLogger(logger)
		}
		logger = logger.With("module", "main")
		context.Config = config
		context.Logger = logger
		return nil
	}
}

// Copy from: "github.com/cosmos/cosmos-sdk/server"
// Justification: function is not exported (private) and used by PersistentPreRunEFn()
// Original description:
//   If a new config is created, change some of the default tendermint settings
func interceptLoadConfig() (conf *tmCfg.Config, err error) {
	tmpConf := tmCfg.DefaultConfig()
	err = viper.Unmarshal(tmpConf)
	if err != nil {
		// TODO: Handle with #870
		panic(err)
	}
	rootDir := tmpConf.RootDir
	configFilePath := filepath.Join(rootDir, "config/config.toml")
	// Intercept only if the file doesn't already exist

	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		// the following parse config is needed to create directories
		conf, _ = tmCmd.ParseConfig() // NOTE: ParseConfig() creates dir/files as necessary.
		conf.ProfListenAddress = "localhost:6060"
		conf.P2P.RecvRate = 5120000
		conf.P2P.SendRate = 5120000
		conf.TxIndex.IndexAllTags = true
		conf.Consensus.TimeoutCommit = 5 * time.Second
		tmCfg.WriteConfigFile(configFilePath, conf)
		// Fall through, just so that its parsed into memory.
	}

	if conf == nil {
		conf, err = tmCmd.ParseConfig() // NOTE: ParseConfig() creates dir/files as necessary.
		if err != nil {
			panic(err)
		}
	}

	appConfigFilePath := filepath.Join(rootDir, "config/app.toml")
	if _, err := os.Stat(appConfigFilePath); os.IsNotExist(err) {
		appConf, _ := sdkSrvCfg.ParseConfig()
		sdkSrvCfg.WriteConfigFile(appConfigFilePath, appConf)
	}

	viper.SetConfigName("app")
	err = viper.MergeInConfig()

	return
}
