package types

import "github.com/dfinance/dnode/helpers/perms"

const (
	// Create a new currency
	PermCCCreator perms.Permission = ModuleName + "PermCCCreator"
	// Update currency supply
	PermCCUpdater perms.Permission = ModuleName + "PermCCUpdater"
	// Read currency and resources
	PermCCReader perms.Permission = ModuleName + "PermCCReader"
	// Update currency VM resources
	PermCCResUpdater perms.Permission = ModuleName + "PermCCResUpdater"
)

var (
	AvailablePermissions = perms.Permissions{PermCCCreator, PermCCUpdater, PermCCReader, PermCCResUpdater}
)

func NewModulePerms() perms.ModulePermissions {
	return perms.NewModulePermissions(ModuleName, AvailablePermissions)
}
