package client

import "github.com/dfinance/dnode/x/orders/internal/types"

const (
	// Permissions
	PermOrderPost   = types.PermOrderPost
	PermOrderRevoke = types.PermOrderRevoke
	PermRead        = types.PermRead
	PermOrderLock   = types.PermOrderLock
	PermOrderUnlock = types.PermOrderUnlock
	PermExecFill    = types.PermExecFill
)
