package core

import (
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	Codespace = "core"
)

var (
	ErrInternal = sdkErrors.Register(Codespace, 100, "internal")
	// StdTx Fee.Amount is empty
	ErrFeeRequired = sdkErrors.Register(Codespace, 101, "tx must contain fees")
	// StdTx Fee.Amount wrong denom
	ErrWrongFeeDenom = sdkErrors.Register(Codespace, 102, "tx must contain fees with a different denom")
	// Module doesn't support multi signature
	ErrNotMultisigModule = sdkErrors.Register(Codespace, 200, "module supports only multisig calls")
)
