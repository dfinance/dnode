package types

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/x/params"
)

const (
	DefaultVMAddress = "0.0.0.0:9671"
)

var (
	KeyVMAddress = []byte("vm_address")
)

type Params struct {
	VMAddress string `json:"vm_address"`
}

func NewParams(vmAddress string) Params {
	return Params{VMAddress: vmAddress}
}

func (p *Params) ParamSetPairs() params.ParamSetPairs {
	return params.ParamSetPairs{
		{Key: KeyVMAddress, Value: &p.VMAddress},
	}
}

func (p Params) Equal(p2 Params) bool {
	return p.VMAddress == p2.VMAddress
}

func (Params) Validate() error {
	return nil
}

func (p Params) String() string {
	return fmt.Sprintf("VMAddress: %s\n", p.VMAddress)
}

func DefaultParams() Params {
	return NewParams(DefaultVMAddress)
}
