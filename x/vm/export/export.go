package export

import "github.com/dfinance/dnode/x/vm/internal/types"

const (
	// Permissions
	PermVmExec        = types.PermVmExec
	PermDsAdmin       = types.PermDsAdmin
	PermStorageReader = types.PermStorageReader
	PermStorageWriter = types.PermStorageWriter
)

type (
	PathData = types.PathData
)
