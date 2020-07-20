// Genesis state for VM block.
package types

// Genesis write operations.
type GenesisWriteOp struct {
	Address string `json:"address" yaml:"address"`
	Path    string `json:"path" yaml:"path"`
	Value   string `json:"value" yaml:"value"`
}

// Genesis state contains write operations.
type GenesisState struct {
	WriteSet []GenesisWriteOp `json:"write_set" yaml:"write_set"`
}
