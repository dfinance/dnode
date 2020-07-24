package types

import (
	"encoding/hex"
	"fmt"

	"github.com/dfinance/dvm-proto/go/vm_grpc"

	"github.com/dfinance/dnode/x/common_vm"
)

// GenesisState is module's genesis (initial state).
type GenesisState struct {
	WriteSet []GenesisWriteOp `json:"write_set" yaml:"write_set"`
}

// Genesis writeSet operation.
type GenesisWriteOp struct {
	Address string `json:"address" yaml:"address"`
	Path    string `json:"path" yaml:"path"`
	Value   string `json:"value" yaml:"value"`
}

func (writeOp GenesisWriteOp) String() string {
	return fmt.Sprintf("%s::%s", writeOp.Address, writeOp.Path)
}

// ToBytes converts GenesisWriteOp to vm_grpc.VMAccessPath and []byte representation for value.
func (writeOp GenesisWriteOp) ToBytes() (*vm_grpc.VMAccessPath, []byte, error) {
	bzAddr, err := hex.DecodeString(writeOp.Address)
	if err != nil {
		return nil, nil, fmt.Errorf("address: %w", err)
	}

	bzPath, err := hex.DecodeString(writeOp.Path)
	if err != nil {
		return nil, nil, fmt.Errorf("path: %w", err)
	}

	bzValue, err := hex.DecodeString(writeOp.Value)
	if err != nil {
		return nil, nil, fmt.Errorf("value: %w", err)
	}

	accessPath := vm_grpc.VMAccessPath{
		Address: bzAddr,
		Path:    bzPath,
	}

	return &accessPath, bzValue, nil
}

// Validate checks that genesis state is valid.
func (s GenesisState) Validate() error {
	writeOpsSet := make(map[string]bool, len(s.WriteSet))
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

		writeOpId := writeOp.String()
		if writeOpsSet[writeOpId] {
			return fmt.Errorf("writeSet[%d]: duplicated %q", woIdx, writeOpId)
		}
		writeOpsSet[writeOpId] = true
	}

	return nil
}

// DefaultGenesisState returns default genesis state (validation is done on module init).
func DefaultGenesisState() GenesisState {
	return GenesisState{}
}
