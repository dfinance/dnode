package types

import (
	"github.com/dfinance/dnode/helpers/perms"
	vmClient "github.com/dfinance/dnode/x/vm/client"
)

const (
	// Init genesis
	PermInit perms.Permission = ModuleName + "PermInit"
	// Read validators and counters
	PermRead perms.Permission = ModuleName + "PermRead"
	// Add/update validators
	PermWrite perms.Permission = ModuleName + "PermWrite"
)

var (
	AvailablePermissions = perms.Permissions{PermInit, PermRead, PermWrite}
)

func NewModulePerms() perms.ModulePermissions {
	return perms.NewModulePermissions(ModuleName, AvailablePermissions)
}

// RequestVMStoragePerms returns module perms used by this module.
func RequestVMStoragePerms() perms.RequestModulePermissions {
	return func() (moduleName string, modulePerms perms.Permissions) {
		moduleName = ModuleName
		modulePerms = perms.Permissions{
			vmClient.PermStorageRead,
			vmClient.PermStorageWrite,
		}
		return
	}
}
