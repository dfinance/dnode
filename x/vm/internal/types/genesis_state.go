// Genesis state for VM block.
package types

import (
	"encoding/hex"
	"fmt"

	"github.com/dfinance/dnode/x/common_vm"
)

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

// Validate checks that genesis state is valid.
func (s GenesisState) Validate() error {
	for woIdx, writeOp := range s.WriteSet {
		bzAddr, err := hex.DecodeString(writeOp.Address)
		if err != nil {
			return fmt.Errorf("writeOp[%d]: address %q: %w", woIdx, writeOp.Address, err)
		}
		if len(bzAddr) != common_vm.VMAddressLength {
			return fmt.Errorf("writeOp[%d]: address %q: incorrect length, should be %d bytes length", woIdx, writeOp.Address, common_vm.VMAddressLength)
		}

		if _, err := hex.DecodeString(writeOp.Path); err != nil {
			return fmt.Errorf("writeOp[%d]: path %q: %w", woIdx, writeOp.Path, err)
		}

		if _, err := hex.DecodeString(writeOp.Value); err != nil {
			return fmt.Errorf("writeSet[%d]: value: %w", woIdx, err)
		}
	}

	return nil
}

// DefaultGenesisState returns default genesis state (validation is done on module init).
func DefaultGenesisState() GenesisState {
	return GenesisState{}
}