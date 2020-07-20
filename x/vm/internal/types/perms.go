package types

import "github.com/dfinance/dnode/helpers/perms"

const (
	// Init genesis, gov operations
	PermInit perms.Permission = ModuleName + "PermInit"
	// Read POA data
	PermVmExec perms.Permission = ModuleName + "PermVmExec"
	// DS start/stop/setContext
	PermDsAdmin perms.Permission = ModuleName + "PermDsAdmin"
	// Read from VM storage
	PermStorageReader perms.Permission = ModuleName + "PermStorageReader"
	// Write to VM storage
	PermStorageWriter perms.Permission = ModuleName + "PermStorageWriter"
)

var (
	AvailablePermissions = perms.Permissions{PermInit, PermVmExec, PermDsAdmin, PermStorageReader, PermStorageWriter}
)

func NewModulePerms() perms.ModulePermissions {
	return perms.NewModulePermissions(ModuleName, AvailablePermissions)
}
