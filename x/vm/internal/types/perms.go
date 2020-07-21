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
	PermStorageRead perms.Permission = ModuleName + "PermStorageRead"
	// Write to VM storage
	PermStorageWrite perms.Permission = ModuleName + "PermStorageWrite"
)

var (
	AvailablePermissions = perms.Permissions{PermInit, PermVmExec, PermDsAdmin, PermStorageRead, PermStorageWrite}
)

func NewModulePerms() perms.ModulePermissions {
	return perms.NewModulePermissions(ModuleName, AvailablePermissions)
}
