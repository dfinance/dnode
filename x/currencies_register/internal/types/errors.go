package types

import (
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrInternal           = sdkErrors.Register(ModuleName, 100, "internal error")
	ErrExists             = sdkErrors.Register(ModuleName, 101, "currency already exists")
	ErrNotFound           = sdkErrors.Register(ModuleName, 102, "currency not found")
	ErrInvalidPath        = sdkErrors.Register(ModuleName, 103, "path is empty")
	ErrWrongCurrencyInfo  = sdkErrors.Register(ModuleName, 104, "wrong currencyInfo params")
	ErrLcsMarshal         = sdkErrors.Register(ModuleName, 105, "can't marshall lcs")
	ErrLcsUnmarshal       = sdkErrors.Register(ModuleName, 106, "unmarshal lcs")
	ErrGovInvalidProposal = sdkErrors.Register(ModuleName, 200, "invalid proposal")
)
