package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const MsgDeployModuleType = "deploy_module"

var _ sdk.Msg = MsgDeployModule{}

// Client message to deploy a module (contract) to VM.
type MsgDeployModule struct {
	Signer sdk.AccAddress `json:"signer" yaml:"signer"`
	Module []Contract     `json:"module" yaml:"module"`
}

// Implements sdk.Msg interface.
func (MsgDeployModule) Route() string {
	return RouterKey
}

// Implements sdk.Msg interface.
func (MsgDeployModule) Type() string {
	return MsgDeployModuleType
}

// Implements sdk.Msg interface.
func (msg MsgDeployModule) ValidateBasic() error {
	if msg.Signer.Empty() {
		return sdkErrors.Wrapf(sdkErrors.ErrInvalidAddress, "empty deployer address")
	}

	if len(msg.Module) == 0 {
		return ErrEmptyContract
	}

	return nil
}

// Implements sdk.Msg interface.
func (msg MsgDeployModule) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// Implements sdk.Msg interface.
func (msg MsgDeployModule) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}

// NewMsgDeployModule creates a new MsgDeployModule message.
func NewMsgDeployModule(signer sdk.AccAddress, modules []Contract) MsgDeployModule {
	return MsgDeployModule{
		Signer: signer,
		Module: modules,
	}
}
