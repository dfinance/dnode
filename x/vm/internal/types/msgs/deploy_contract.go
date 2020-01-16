package msgs

import (
	"encoding/json"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"wings-blockchain/x/vm/internal/types"
)

var (
	MsgDeployContractType = types.ModuleName + "/MsgDeployContract"

	_ sdk.Msg = MsgDeployContract{}
)

type MsgDeployContract struct {
	Signer   sdk.AccAddress `json:"signer"`
	Contract types.Contract `json:"contract"`
}

func (MsgDeployContract) Route() string {
	return types.RouteKey
}

func (MsgDeployContract) Type() string {
	return "deploy_contract"
}

func (msg MsgDeployContract) ValidateBasic() sdk.Error {
	if msg.Signer.Empty() {
		return sdk.ErrInvalidAddress("deployer address is empty")
	}

	if len(msg.Contract) == 0 {
		return types.ErrEmptyContract()
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
