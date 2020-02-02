package types

import (
	"encoding/hex"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	CodeEmptyContractCode = 101

	CodeErrDuringExec = 201

	CodeErrWrongModuleAddress = 301
	CodeErrModuleExists       = 302
	CodeErrWrongAddressLength = 303
	CodeErrWrongArgTypeTag    = 304
)

type ErrVMCrashed struct {
	err error
}

func NewErrVMCrashed(err error) ErrVMCrashed {
	return ErrVMCrashed{err: err}
}

func ErrEmptyContract() sdk.Error {
	return sdk.NewError(Codespace, CodeEmptyContractCode, "contract code is empty, please fill field with compiled contract bytes")
}

func ErrDuringVMExec(msg string) sdk.Error {
	return sdk.NewError(Codespace, CodeErrDuringExec, "can't execute contract: %s", msg)
}

func ErrWrongModuleAddress(expected, real sdk.AccAddress) sdk.Error {
	return sdk.NewError(Codespace, CodeErrWrongModuleAddress, "wrong module owner %s address, expected %s", expected, real)
}

func ErrModuleExists(address sdk.AccAddress, path []byte) sdk.Error {
	return sdk.NewError(Codespace, CodeErrModuleExists, "module %s already exists for account %s", hex.EncodeToString(path), address)
}

func ErrWrongAddressLength(address sdk.AccAddress) sdk.Error {
	return sdk.NewError(Codespace, CodeErrWrongAddressLength, "address %s passed to vm has wrong length, it has length %d, but expected %d", address.String(), len(address), VmAddressLength)
}

func ErrWrongArgTypeTag(err error) sdk.Error {
	return sdk.NewError(Codespace, CodeErrWrongArgTypeTag, "something wrong with argument type: %s", err.Error())
}
