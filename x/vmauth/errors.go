package vmauth

import (
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

var (
	ErrInternal = sdkErrors.Register(auth.ModuleName, 100, "internal")
)
