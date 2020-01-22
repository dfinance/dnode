// Configuration for WBD and WBCli.
package config

import (
	"bytes"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/viper"
	cmn "github.com/tendermint/tendermint/libs/common"
	"os"
	"path/filepath"
)

const (
	MainDenom            = "wings"
	MainPrefix           = "wallets"                                                                 // Main prefix for all addresses.
	Bech32PrefixAccAddr  = MainPrefix                                                                // Bech32 prefix for account addresses.
	Bech32PrefixAccPub   = MainPrefix + sdk.PrefixPublic                                             // Bech32 prefix for accounts pub keys.
	Bech32PrefixValAddr  = MainPrefix + sdk.PrefixValidator + sdk.PrefixOperator                     // Bech32 prefix for validators addresses.
	Bech32PrefixValPub   = MainPrefix + sdk.PrefixValidator + sdk.PrefixOperator + sdk.PrefixPublic  // Bech32 prefix for validator pub keys.
	Bech32PrefixConsAddr = MainPrefix + sdk.PrefixValidator + sdk.PrefixConsensus                    // Bech32 prefix for consensus addresses.
	Bech32PrefixConsPub  = MainPrefix + sdk.PrefixValidator + sdk.PrefixConsensus + sdk.PrefixPublic // Bech32 prefix for consensus pub keys.

	VMConfigFile            = "vm.toml"
	ConfigDir               = "config"
	DefaultVMAddress        = "127.0.0.1:50051"
	DefaultVMRequestTimeout = 100
)

// Virtual machine connection config
type VMConfig struct {
	Address       string `mapstructure:"vm_address"`
	DeployTimeout uint64 `mapstructure:"vm_deploy_timeout"`
}

func DefaultVMConfig() *VMConfig {
	return &VMConfig{
		Address:       DefaultVMAddress,
		DeployTimeout: DefaultVMRequestTimeout,
	}
}

// Initializing WB custom prefixes.
func InitBechPrefixes(config *sdk.Config) {
	config.SetBech32PrefixForAccount(Bech32PrefixAccAddr, Bech32PrefixAccPub)
	config.SetBech32PrefixForValidator(Bech32PrefixValAddr, Bech32PrefixValPub)
	config.SetBech32PrefixForConsensusNode(Bech32PrefixConsAddr, Bech32PrefixConsPub)
}

func WriteVMConfig(rootDir string, vmConfig *VMConfig) {
	configFilePath := filepath.Join(rootDir, VMConfigFile)

	var buffer bytes.Buffer

	if err := configTemplate.Execute(&buffer, vmConfig); err != nil {
		panic(err)
	}

	cmn.MustWriteFile(configFilePath, buffer.Bytes(), 0644)
}

func ReadVMConfig(rootDir string) (*VMConfig, error) {
	configFilePath := filepath.Join(rootDir, ConfigDir, VMConfigFile)

	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		config := DefaultVMConfig()
		WriteVMConfig(rootDir, config)

		return config, nil
	}

	viper.SetConfigFile(configFilePath)

	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	// read config
	config := DefaultVMConfig()
	if err := viper.Unmarshal(config); err != nil {
		panic(err)
	}

	return config, nil
}
