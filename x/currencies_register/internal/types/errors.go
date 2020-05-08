package types

import (
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrInternal        = sdkErrors.Register(ModuleName, 100, "internal error")
	ErrExists          = sdkErrors.Register(ModuleName, 101, "currency already exists")
	ErrWrongAddressLen = sdkErrors.Register(ModuleName, 102, "wrong length address")
	ErrLcsMarshal      = sdkErrors.Register(ModuleName, 103, "cant marshall lcs")
	ErrNotFound        = sdkErrors.Register(ModuleName, 104, "currency not found")
	ErrLcsUnmarshal    = sdkErrors.Register(ModuleName, 105, "unmarshal lcs")
)
