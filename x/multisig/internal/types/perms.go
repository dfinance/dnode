package types

import (
	"github.com/dfinance/dnode/helpers/perms"
)

const (
	// Init genesis
	PermInit perms.Permission = ModuleName + "PermInit"
	// Read calls, handlers
	PermRead perms.Permission = ModuleName + "PermRead"
	// Add/update calls
	PermWrite perms.Permission = ModuleName + "PermWrite"
)

var (
	AvailablePermissions = perms.Permissions{PermInit, PermRead, PermWrite}
)

func NewModulePerms() perms.ModulePermissions {
	return perms.NewModulePermissions(ModuleName, AvailablePermissions)
}
