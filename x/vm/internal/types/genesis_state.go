// Genesis state for VM block.
package types

// Genesis write operations.
type GenesisWriteOp struct {
	Address string `json:"address"`
	Path    string `json:"path"`
	Value   string `json:"value"`
}

// Genesis state contains write operations.
type GenesisState struct {
	WriteSet []GenesisWriteOp `json:"write_set"`
}
