package types

import (
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrInternal    = sdkErrors.Register(ModuleName, 100, "internal")
	ErrWrongDenom  = sdkErrors.Register(ModuleName, 101, "wrong denom")
	ErrWrongParams = sdkErrors.Register(ModuleName, 102, "invalid currency params")
)
