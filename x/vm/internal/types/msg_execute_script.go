package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/dfinance/dvm-proto/go/types_grpc"
)

const MsgExecuteScriptType = "execute_script"

var _ sdk.Msg = MsgExecuteScript{}

// Client message to deploy a script with args to VM.
type MsgExecuteScript struct {
	Signer sdk.AccAddress `json:"signer" yaml:"signer"`
	Script Contract       `json:"script" yaml:"script"`
	Args   []ScriptArg    `json:"args" yaml:"args"`
}

// Implements sdk.Msg interface.
func (MsgExecuteScript) Route() string {
	return RouterKey
}

// Implements sdk.Msg interface.
func (MsgExecuteScript) Type() string {
	return MsgExecuteScriptType
}

// Implements sdk.Msg interface.
func (msg MsgExecuteScript) ValidateBasic() error {
	if msg.Signer.Empty() {
		return sdkErrors.Wrap(sdkErrors.ErrInvalidAddress, "empty deployer address")
	}

	if len(msg.Script) == 0 {
		return ErrEmptyContract
	}

	for _, arg := range msg.Args {
		if _, err := StringifyVMTypeTag(arg.Type); err != nil {
			return sdkErrors.Wrap(ErrWrongArgTypeTag, err.Error())
		}
		if len(arg.Value) == 0 {
			return ErrWrongArgValue
		}
	}

	return nil
}

// Implements sdk.Msg interface.
func (msg MsgExecuteScript) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(msg))
}

// Implements sdk.Msg interface.
func (msg MsgExecuteScript) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}

// NewMsgExecuteScript creates a new MsgExecuteScript message.
func NewMsgExecuteScript(signer sdk.AccAddress, script Contract, args []ScriptArg) MsgExecuteScript {
	return MsgExecuteScript{
		Signer: signer,
		Script: script,
		Args:   args,
	}
}

// ScriptArg defines VM script argument.
type ScriptArg struct {
	Type  types_grpc.VMTypeTag `json:"type" yaml:"type"`
	Value []byte               `json:"value" yaml:"value"`
}

// NewScriptArg creates a new ScriptArg object.
func NewScriptArg(typeTag types_grpc.VMTypeTag, value []byte) ScriptArg {
	return ScriptArg{
		Type:  typeTag,
		Value: value,
	}
}
