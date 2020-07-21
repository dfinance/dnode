package types

import (
	"github.com/dfinance/dnode/helpers/perms"
	ccsClient "github.com/dfinance/dnode/x/ccstorage/client"
)

const (
	// Init genesis
	PermInit perms.Permission = ModuleName + "PermInit"
	// Create a new market / modify params
	PermCreate perms.Permission = ModuleName + "PermCreate"
	// Read market / markets
	PermRead perms.Permission = ModuleName + "PermRead"
)

var (
	AvailablePermissions = perms.Permissions{PermInit, PermCreate, PermRead}
)

func NewModulePerms() perms.ModulePermissions {
	return perms.NewModulePermissions(ModuleName, AvailablePermissions)
}

// RequestCCStoragePerms returns module perms used by this module.
func RequestCCStoragePerms() perms.RequestModulePermissions {
	return func() (moduleName string, modulePerms perms.Permissions) {
		moduleName = ModuleName
		modulePerms = perms.Permissions{
			ccsClient.PermRead,
		}
		return
	}
}
