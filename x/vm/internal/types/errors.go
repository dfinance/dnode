package types

import sdk "github.com/cosmos/cosmos-sdk/types"

const (
	CodeEmptyContractCode = 101

	CodeCantConnectVM = 201
	CodeErrDuringExec = 202
)

func ErrEmptyContract() sdk.Error {
	return sdk.NewError(Codespace, CodeEmptyContractCode, "contract code is empty, please fill field with compiled contract bytes")
}

func ErrCantConnectVM(msg string) sdk.Error {
	return sdk.NewError(Codespace, CodeCantConnectVM, "cant connect to vm instance: %s", msg)
}

func ErrDuringVMExec(msg string) sdk.Error {
	return sdk.NewError(Codespace, CodeErrDuringExec, "cant execute contract: %s", msg)
}
