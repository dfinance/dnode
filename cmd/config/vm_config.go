package config

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	tmOs "github.com/tendermint/tendermint/libs/os"
)

const (
	VMConfigFile = "vm.toml" // Default file to store config.
	ConfigDir    = "config"  // Default directory to store all configurations.

	// VM configs.
	DefaultVMAddress    = "tcp://127.0.0.1:50051" // Default virtual machine address to connect from Cosmos SDK.
	DefaultDataListen   = "tcp://127.0.0.1:50052" // Default data server address to listen for connections from VM.
	DefaultCompilerAddr = DefaultVMAddress

	// Default retry configs.
	DefaultMaxAttempts = 0 // Default maximum attempts for retry.
	DefaultReqTimeout  = 0 // Default request timeout per attempt [ms].

	// Invariants check period for crisis module (in blocks)
	DefInvCheckPeriod = 10
)

// Virtual machine connection config (see config/vm.toml).
type VMConfig struct {
	Address    string `mapstructure:"vm_address"`     // address of virtual machine.
	DataListen string `mapstructure:"vm_data_listen"` // data listen.

	// Retry policy
	MaxAttempts    uint `mapstructure:"vm_retry_max_attempts"`   // maximum attempts for retry (0 - infinity)
	ReqTimeoutInMs uint `mapstructure:"vm_retry_req_timeout_ms"` // request timeout per attempt (0 - infinity) [ms]
}

// Default VM configuration.
func DefaultVMConfig() *VMConfig {
	return &VMConfig{
		Address:        DefaultVMAddress,
		DataListen:     DefaultDataListen,
		MaxAttempts:    DefaultMaxAttempts,
		ReqTimeoutInMs: DefaultReqTimeout,
	}
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
