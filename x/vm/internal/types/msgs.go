package types

import (
	"encoding/json"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	MsgDeployContractType = ModuleName + "/MsgDeployContract"

	_ sdk.Msg = MsgDeployContract{}
)

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
