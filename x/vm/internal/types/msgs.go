package types

import (
	"encoding/json"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	MsgDeployContractType = ModuleName + "/MsgDeployContract"
	MsgScriptContractType = ModuleName + "/MsgScriptContract"

	_ sdk.Msg = MsgDeployContract{}
	_ sdk.Msg = MsgScriptContract{}
)

// Message to deploy contract.
type MsgDeployContract struct {
	Signer   sdk.AccAddress `json:"signer"`
	Contract Contract       `json:"contract"`
}

func NewMsgDeployContract(signer sdk.AccAddress, contract Contract) MsgDeployContract {
	return MsgDeployContract{
		Signer:   signer,
		Contract: contract,
	}
}

func (MsgDeployContract) Route() string {
	return RouterKey
}

func (MsgDeployContract) Type() string {
	return "deploy_contract"
}

func (msg MsgDeployContract) ValidateBasic() sdk.Error {
	if msg.Signer.Empty() {
		return sdk.ErrInvalidAddress("deployer address is empty")
	}

	if len(msg.Contract) == 0 {
		return ErrEmptyContract()
	}

	return nil
}

func (msg MsgDeployContract) GetSignBytes() []byte {
	bc, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}

	return sdk.MustSortJSON(bc)
}

func (msg MsgDeployContract) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}

// Message for contract script (execution).
type MsgScriptContract struct {
	Signer   sdk.AccAddress `json:"signer"`
	Contract Contract       `json:"contract"`
}

func NewMsgScriptContract(signer sdk.AccAddress, contract Contract) MsgScriptContract {
	return MsgScriptContract{
		Signer:   signer,
		Contract: contract,
	}
}

func (MsgScriptContract) Route() string {
	return RouterKey
}

func (MsgScriptContract) Type() string {
	return "script_contract"
}

func (msg MsgScriptContract) ValidateBasic() sdk.Error {
	if msg.Signer.Empty() {
		return sdk.ErrInvalidAddress("deployer address is empty")
	}

	if len(msg.Contract) == 0 {
		return ErrEmptyContract()
	}

	return nil
}

func (msg MsgScriptContract) GetSignBytes() []byte {
	bc, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}

	return sdk.MustSortJSON(bc)
}

func (msg MsgScriptContract) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{msg.Signer}
}
