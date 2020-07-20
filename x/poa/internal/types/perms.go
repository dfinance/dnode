package types

import "github.com/dfinance/dnode/helpers/perms"

const (
	// Init genesis
	PermInit perms.Permission = ModuleName + "PermInit"
	// Read validators and counters
	PermReader perms.Permission = ModuleName + "PermReader"
	// Add/update validators
	PermWriter perms.Permission = ModuleName + "PermWriter"
)

var (
	AvailablePermissions = perms.Permissions{PermInit, PermReader, PermWriter}
)

func NewModulePerms() perms.ModulePermissions {
	return perms.NewModulePermissions(ModuleName, AvailablePermissions)
}
