// Configuration for DNode and DNCli.
package config

import (
	"bytes"
	"os"
	"path/filepath"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/viper"
	tmOs "github.com/tendermint/tendermint/libs/os"
)

const (
	MainDenom            = "dfi"
	DefaultFee           = "1" + MainDenom
	MainPrefix           = "wallet"                                                                  // Main prefix for all addresses.
	Bech32PrefixAccAddr  = MainPrefix                                                                // Bech32 prefix for account addresses.
	Bech32PrefixAccPub   = MainPrefix + sdk.PrefixPublic                                             // Bech32 prefix for accounts pub keys.
	Bech32PrefixValAddr  = MainPrefix + sdk.PrefixValidator + sdk.PrefixOperator                     // Bech32 prefix for validators addresses.
	Bech32PrefixValPub   = MainPrefix + sdk.PrefixValidator + sdk.PrefixOperator + sdk.PrefixPublic  // Bech32 prefix for validator pub keys.
	Bech32PrefixConsAddr = MainPrefix + sdk.PrefixValidator + sdk.PrefixConsensus                    // Bech32 prefix for consensus addresses.
	Bech32PrefixConsPub  = MainPrefix + sdk.PrefixValidator + sdk.PrefixConsensus + sdk.PrefixPublic // Bech32 prefix for consensus pub keys.

	VMConfigFile = "vm.toml" // Default file to store config.
	ConfigDir    = "config"  // Default directory to store all configurations.

	// VM configs.
	DefaultVMAddress  = "tcp://127.0.0.1:50051" // Default virtual machine address to connect from Cosmos SDK.
	DefaultDataListen = "tcp://127.0.0.1:50052" // Default data server address to listen for connections from VM.

	// Default retry configs.
	DefaultMaxAttempts       = 0    // Default VM retry attempts.
	DefaultInitialBackoff    = 1000 // Default VM 100 milliseconds for retry attempts.
	DefaultMaxBackoff        = 2000 // Default VM max backoff.
	DefaultBackoffMultiplier = 0.1  // Default backoff multiplayer (10)

	// Default governance params.
	DefaultGovMinDepositAmount = "100000000000000000000" // 100 dfi
)

var (
	GovMinDeposit sdk.Coin
)

// Virtual machine connection config (see config/vm.toml).
type VMConfig struct {
	Address    string `mapstructure:"vm_address"`     // address of virtual machine.
	DataListen string `mapstructure:"vm_data_listen"` // data listen.

	// Retry policy.
	// Example how backoff works - https://stackoverflow.com/questions/43224683/what-does-backoffmultiplier-mean-in-defaultretrypolicy.
	MaxAttempts       int     `mapstructure:"vm_retry_max_attempts"`       // maximum attempts for retry, for infinity retry - use 0.
	InitialBackoff    int     `mapstructure:"vm_retry_initial_backoff"`    // initial back off in ms.
	MaxBackoff        int     `mapstructure:"vm_retry_max_backoff"`        // max backoff in ms.
	BackoffMultiplier float64 `mapstructure:"vm_retry_backoff_multiplier"` // backoff multiplier.
}

// Default VM configuration.
func DefaultVMConfig() *VMConfig {
	return &VMConfig{
		Address:           DefaultVMAddress,
		DataListen:        DefaultDataListen,
		MaxAttempts:       DefaultMaxAttempts,
		InitialBackoff:    DefaultInitialBackoff,
		MaxBackoff:        DefaultMaxBackoff,
		BackoffMultiplier: DefaultBackoffMultiplier,
	}
}

// Initializing DN custom prefixes.
func InitBechPrefixes(config *sdk.Config) {
	config.SetBech32PrefixForAccount(Bech32PrefixAccAddr, Bech32PrefixAccPub)
	config.SetBech32PrefixForValidator(Bech32PrefixValAddr, Bech32PrefixValPub)
	config.SetBech32PrefixForConsensusNode(Bech32PrefixConsAddr, Bech32PrefixConsPub)
}

// Write VM config file in configuration directory.
func WriteVMConfig(rootDir string, vmConfig *VMConfig) {
	configFilePath := filepath.Join(rootDir, ConfigDir, VMConfigFile)

	var buffer bytes.Buffer

	if err := configTemplate.Execute(&buffer, vmConfig); err != nil {
		panic(err)
	}

	tmOs.MustWriteFile(configFilePath, buffer.Bytes(), 0644)
}

// Read VM config file from configuration directory.
func ReadVMConfig(rootDir string) (*VMConfig, error) {
	configFilePath := filepath.Join(rootDir, ConfigDir, VMConfigFile)

	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		config := DefaultVMConfig()
		WriteVMConfig(rootDir, config)
		return config, nil
	}

	viper.SetConfigFile(configFilePath)

	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}

	// read config
	config := DefaultVMConfig()
	if err := viper.Unmarshal(config); err != nil {
		panic(err)
	}

	return config, nil
}

func init() {
	minDepositAmount, ok := sdk.NewIntFromString(DefaultGovMinDepositAmount)
	if !ok {
		panic("governance genesisState: minDeposit convertation failed")
	}

	GovMinDeposit = sdk.NewCoin(MainDenom, minDepositAmount)
}
