package types

import (
	"github.com/dfinance/dnode/helpers/perms"
	"github.com/dfinance/dnode/x/ccstorage"
)

const (
	// Init genesis
	PermInit perms.Permission = ModuleName + "PermInit"
	// Create a new market / modify params
	PermCreator perms.Permission = ModuleName + "PermCreator"
	// Read market / markets
	PermReader perms.Permission = ModuleName + "PermReader"
)

var (
	AvailablePermissions = perms.Permissions{PermInit, PermCreator, PermReader}
)

func NewModulePerms() perms.ModulePermissions {
	return perms.NewModulePermissions(ModuleName, AvailablePermissions)
}

// RequestCCStoragePerms returns module perms used by this module.
func RequestCCStoragePerms() perms.RequestModulePermissions {
	return func() (moduleName string, modulePerms perms.Permissions) {
		moduleName = ModuleName
		modulePerms = perms.Permissions{
			ccstorage.PermCCReader,
		}
		return
	}
}
