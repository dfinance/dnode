package types

import (
	"github.com/dfinance/dnode/helpers/perms"
	vmClient "github.com/dfinance/dnode/x/vm/client"
)

const (
	// Init genesis
	PermInit perms.Permission = ModuleName + "PermInit"
	// Create a new currency
	PermCreate perms.Permission = ModuleName + "PermCreate"
	// Update currency supply
	PermUpdate perms.Permission = ModuleName + "PermUpdate"
	// Read currency and resources
	PermRead perms.Permission = ModuleName + "PermRead"
	// Update currency VM resources
	PermResUpdate perms.Permission = ModuleName + "PermResUpdate"
)

var (
	AvailablePermissions = perms.Permissions{PermInit, PermCreate, PermUpdate, PermRead, PermResUpdate}
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
