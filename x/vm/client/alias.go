package client

import "github.com/dfinance/dnode/x/vm/internal/types"

type (
	PathData = types.PathData
)

const (
	// Permissions
	PermVmExec       = types.PermVmExec
	PermDsAdmin      = types.PermDsAdmin
	PermStorageRead  = types.PermStorageRead
	PermStorageWrite = types.PermStorageWrite
)
