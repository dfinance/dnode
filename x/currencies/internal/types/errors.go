package types

import (
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	ErrInternal            = sdkErrors.Register(ModuleName, 100, "internal")
	ErrWrongDenom          = sdkErrors.Register(ModuleName, 101, "wrong denom")
	ErrWrongAmount         = sdkErrors.Register(ModuleName, 102, "wrong amount")
	ErrWrongIssueID        = sdkErrors.Register(ModuleName, 103, "wrong issueID")
	ErrWrongWithdrawID     = sdkErrors.Register(ModuleName, 104, "wrong withdrawID")
	ErrWrongPegZoneSpender = sdkErrors.Register(ModuleName, 105, "wrong PegZone spender")
	ErrGovInvalidProposal  = sdkErrors.Register(ModuleName, 200, "invalid proposal")
)
