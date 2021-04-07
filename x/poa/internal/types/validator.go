package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/dfinance/dnode/helpers"
)

// Validator is a PoA validator meta.
type Validator struct {
	// SDK address
	Address sdk.AccAddress `json:"address" yaml:"address" example:"wallet1a7280dyzp487r7wghr99f6r3h2h2z4gk4d740m"`
	// Ethereum address
	EthAddress string `json:"eth_address" yaml:"eth_address" example:"0x29D7d1dd5B6f9C864d9db560D72a247c178aE86B"`
}

// Validate checks Validator.
func (v Validator) Validate() error {
	if v.Address.Empty() {
		return sdkErrors.Wrap(sdkErrors.ErrInvalidAddress, "address: empty")
	}

	if len(v.EthAddress) == 0 {
		return sdkErrors.Wrap(ErrWrongEthereumAddress, "ethAddress: empty")
	}
	if !helpers.IsEthereumAddress(v.EthAddress) {
		return sdkErrors.Wrapf(ErrWrongEthereumAddress, "ethAddress: invalid %s for %s", v.EthAddress, v.Address.String())
	}

	return nil
}

func (v Validator) String() string {
	return fmt.Sprintf("Validator:\n"+
		"Address: %s\n"+
		"EthAddress: %s",
		v.Address.String(),
		v.EthAddress,
	)
}

// NewValidator creates a new Validator.
func NewValidator(address sdk.AccAddress, ethAddress string) Validator {
	return Validator{
		Address:    address,
		EthAddress: ethAddress,
	}
}

// Slice of Validator objects.
type Validators []Validator

func (list Validators) String() string {
	strBuilder := strings.Builder{}
	for i, v := range list {
		strBuilder.WriteString(fmt.Sprintf("[%d] %s", i, v.String()))
		if i < len(list)-1 {
			strBuilder.WriteString("\n")
		}
	}

	return strBuilder.String()
}
