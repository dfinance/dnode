package types

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/x/params"
	"time"
)

const (
	DefaultVMAddress = "0.0.0.0:50051"
	DefaultVMTimeout = 100
)

var (
	KeyVMAddress = []byte("vmaddress")
	KeyVMTimeout = []byte("vmtimeout")
)

type Params struct {
	VMAddress string        `json:"vm_address"` // Address to connect to VM via grpc
	VMTimeout time.Duration `json:"vm_timeout"` // VM timeout in milliseconds.
}

func NewParams(vmAddress string, vmTimeout time.Duration) Params {
	return Params{VMAddress: vmAddress, VMTimeout: vmTimeout}
}

func (p *Params) ParamSetPairs() params.ParamSetPairs {
	return params.ParamSetPairs{
		{Key: KeyVMAddress, Value: &p.VMAddress},
		{Key: KeyVMTimeout, Value: &p.VMTimeout},
	}
}

func (p Params) Equal(p2 Params) bool {
	return p.VMAddress == p2.VMAddress && p.VMTimeout == p2.VMTimeout
}

func (Params) Validate() error {
	return nil
}

func (p Params) String() string {
	return fmt.Sprintf("VMAddress: %s\n", p.VMAddress)
}

func DefaultParams() Params {
	return NewParams(DefaultVMAddress, DefaultVMTimeout)
}
